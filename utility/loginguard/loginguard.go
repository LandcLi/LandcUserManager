// Package loginguard 提供登录失败计数与账号锁定。
//
// 设计要点：
//   - **按标识符（邮箱/手机号/用户名）计数**，不按用户 ID —— 否则「猜不存在的
//     用户名」这条路完全不受限制，攻击者可以无限次探测账号是否存在。
//   - 计数与锁定状态存放在框架缓存（有 Redis 用 Redis、无 Redis 自动降级为
//     进程内缓存）。多实例部署只有配了 Redis 才是全局一致的，这一点写在 README 里，
//     不做「看起来分布式其实各自为政」的假象。
//   - 缓存未初始化时退化为进程内计数器，保证锁定策略不会静默失效。
//   - 锁定是「时间锁」而非「人工解锁」：到期自动恢复，避免攻击者用错误密码
//     把任意已知账号永久锁死（拒绝服务）。
package loginguard

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"

	"github.com/LandcLi/LandcUserManager/utility/config"
)

const (
	// failKeyPrefix 失败计数 key 前缀，值为连续失败次数。
	failKeyPrefix = "lum:login:fail:"
	// lockKeyPrefix 锁定 key 前缀，值为锁定到期的 Unix 秒。
	lockKeyPrefix = "lum:login:lock:"
)

// Policy 是一次判定使用的登录安全策略。
type Policy struct {
	Enabled           bool
	MaxFailedAttempts int
	LockDuration      time.Duration
	FailureWindow     time.Duration
}

// CurrentPolicy 从全局配置读取当前策略。
// 配置未初始化时返回「关闭」策略 —— 宁可不锁，也不能 panic 在登录路径上。
func CurrentPolicy() Policy {
	cfg := config.TryGetConfig()
	if cfg == nil {
		return Policy{}
	}
	p := cfg.LoginSecurity
	lockMinutes := p.LockMinutes
	if lockMinutes <= 0 {
		lockMinutes = 15
	}
	windowMinutes := p.FailureWindowMinutes
	if windowMinutes <= 0 {
		windowMinutes = 15
	}
	return Policy{
		Enabled:           p.Enabled && p.MaxFailedAttempts > 0,
		MaxFailedAttempts: p.MaxFailedAttempts,
		LockDuration:      time.Duration(lockMinutes) * time.Minute,
		FailureWindow:     time.Duration(windowMinutes) * time.Minute,
	}
}

// Normalize 归一化标识符，保证「Alice@Example.com」与「alice@example.com」
// 命中同一份计数（否则改个大小写就能绕过锁定）。
func Normalize(identifier string) string {
	return strings.ToLower(strings.TrimSpace(identifier))
}

// Locked 报告该标识符是否处于锁定状态，并返回剩余锁定时间。
func Locked(ctx context.Context, identifier string) (bool, time.Duration) {
	p := CurrentPolicy()
	id := Normalize(identifier)
	if !p.Enabled || id == "" {
		return false, 0
	}

	raw, ok := storeGet(ctx, lockKeyPrefix+id)
	if !ok || raw == "" {
		return false, 0
	}

	deadline, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		// 值损坏（人工改了缓存等）：按「整段锁定时长」处理，
		// 宁可多锁一会儿也不要因解析失败把锁定直接放行。
		return true, p.LockDuration
	}
	remaining := time.Until(time.Unix(deadline, 0))
	if remaining <= 0 {
		return false, 0
	}
	return true, remaining
}

// OnFailure 记录一次登录失败。
// 达到阈值时写入锁定并返回 (true, 锁定时长)，否则返回 (false, 0)。
func OnFailure(ctx context.Context, identifier string) (bool, time.Duration) {
	p := CurrentPolicy()
	id := Normalize(identifier)
	if !p.Enabled || id == "" {
		return false, 0
	}

	n, err := storeIncr(ctx, failKeyPrefix+id, p.FailureWindow)
	if err != nil {
		// 计数失败不能反过来阻断正常登录：放行，由日志暴露缓存故障。
		return false, 0
	}
	if n < int64(p.MaxFailedAttempts) {
		return false, 0
	}

	deadline := time.Now().Add(p.LockDuration)
	storeSet(ctx, lockKeyPrefix+id, strconv.FormatInt(deadline.Unix(), 10), p.LockDuration)
	// 锁定已经生效，计数归零，避免锁定期满后失败次数仍然超标导致立即再次锁定。
	storeDelete(ctx, failKeyPrefix+id)
	return true, p.LockDuration
}

// OnSuccess 登录成功后清零失败计数并解除锁定。
func OnSuccess(ctx context.Context, identifier string) {
	id := Normalize(identifier)
	if id == "" {
		return
	}
	storeDelete(ctx, failKeyPrefix+id, lockKeyPrefix+id)
}

// Unlock 供改密/重置密码等「证明过身份」的流程以及运维调用：
// 解除锁定并清零计数。identifier 为空时忽略。
func Unlock(ctx context.Context, identifiers ...string) {
	var keys []string
	for _, identifier := range identifiers {
		id := Normalize(identifier)
		if id == "" {
			continue
		}
		keys = append(keys, failKeyPrefix+id, lockKeyPrefix+id)
	}
	if len(keys) == 0 {
		return
	}
	storeDelete(ctx, keys...)
}

// ---------- 存储层：优先框架缓存，未初始化时退化为进程内计数 ----------

// localStore 是框架缓存不可用时的兜底实现。
// 键为完整 key，值带过期时间；只在单实例进程内有效。
var (
	localMu     sync.Mutex
	localStore  = map[string]localEntry{}
	localSweeps int // 每 N 次写入顺带清理一次过期项，避免无界增长
)

type localEntry struct {
	value    string
	expireAt time.Time
}

func (e localEntry) expired(now time.Time) bool {
	return !e.expireAt.IsZero() && now.After(e.expireAt)
}

func storeGet(ctx context.Context, key string) (string, bool) {
	if c := framecache.GetCache(); c != nil {
		v, err := c.Get(ctx, key)
		if err != nil {
			return "", false
		}
		return v, true
	}

	localMu.Lock()
	defer localMu.Unlock()
	e, ok := localStore[key]
	if !ok || e.expired(time.Now()) {
		return "", false
	}
	return e.value, true
}

func storeSet(ctx context.Context, key, value string, ttl time.Duration) {
	if c := framecache.GetCache(); c != nil {
		_ = c.Set(ctx, key, value, ttl)
		return
	}

	localMu.Lock()
	defer localMu.Unlock()
	localStore[key] = localEntry{value: value, expireAt: time.Now().Add(ttl)}
	sweepLocalLocked()
}

func storeIncr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if c := framecache.GetCache(); c != nil {
		return c.Incr(ctx, key, ttl)
	}

	localMu.Lock()
	defer localMu.Unlock()
	now := time.Now()
	e := localStore[key]
	var n int64
	if !e.expired(now) {
		n, _ = strconv.ParseInt(e.value, 10, 64)
	}
	n++
	// 过期时间只在第一次失败时设置（固定窗口），
	// 与 RedisCache.Incr 的语义保持一致。
	expireAt := e.expireAt
	if e.expired(now) || expireAt.IsZero() {
		expireAt = now.Add(ttl)
	}
	localStore[key] = localEntry{value: strconv.FormatInt(n, 10), expireAt: expireAt}
	sweepLocalLocked()
	return n, nil
}

func storeDelete(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if c := framecache.GetCache(); c != nil {
		_ = c.Delete(ctx, keys...)
		return
	}

	localMu.Lock()
	defer localMu.Unlock()
	for _, key := range keys {
		delete(localStore, key)
	}
}

// sweepLocalLocked 每写入 256 次清理一次过期项。调用方需持有 localMu。
func sweepLocalLocked() {
	localSweeps++
	if localSweeps%256 != 0 {
		return
	}
	now := time.Now()
	for key, e := range localStore {
		if e.expired(now) {
			delete(localStore, key)
		}
	}
}

// ResetLocalStore 清空进程内兜底存储。仅用于测试隔离。
func ResetLocalStore() {
	localMu.Lock()
	defer localMu.Unlock()
	localStore = map[string]localEntry{}
	localSweeps = 0
}
