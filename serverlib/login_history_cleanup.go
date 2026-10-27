package serverlib

import (
	"context"
	"sync"
	"time"

	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/log/facade"

	"github.com/LandcLi/LandcUserManager/dao"
)

// 登录历史的后台清理。
//
// 缺陷背景：config 里一直有 login_history.retention_days（示例值 90），
// DAO 也实现了 DeleteOldRecords，但**没有任何地方调用它** ——
// 每次登录（成功或失败）都往 login_histories 插一条，只写不删。
// 长期运行表会无限膨胀，最终拖慢「查最近一次登录」这类高频查询
// （风控要用它判断异地/新设备）。
//
// 与 LandcSSO 的 tokens 清理同源，都是「有配置、有方法、缺调度」。

// loginHistoryCleanupInterval 清理周期。登录历史变化缓慢，一天一次足够。
const loginHistoryCleanupInterval = 24 * time.Hour

var (
	loginHistoryCleanupOnce sync.Once
	loginHistoryCleanupStop chan struct{}
)

// startLoginHistoryCleanup 启动后台清理（幂等）。
//
// retentionDays <= 0 表示关闭：保留全部历史（由运维自行归档）。
func startLoginHistoryCleanup(retentionDays int) {
	if retentionDays <= 0 {
		facade.Warnf("login_history.retention_days <= 0，登录历史清理已关闭，表将持续增长")
		return
	}
	loginHistoryCleanupOnce.Do(func() {
		loginHistoryCleanupStop = make(chan struct{})
		go loginHistoryCleanupLoop(retentionDays, loginHistoryCleanupStop)
	})
}

// stopLoginHistoryCleanup 停止后台清理（测试与优雅停机用，幂等）。
func stopLoginHistoryCleanup() {
	if loginHistoryCleanupStop == nil {
		return
	}
	select {
	case <-loginHistoryCleanupStop:
		// 已停止
	default:
		close(loginHistoryCleanupStop)
	}
}

func loginHistoryCleanupLoop(retentionDays int, stop chan struct{}) {
	ticker := time.NewTicker(loginHistoryCleanupInterval)
	defer ticker.Stop()

	// 启动即执行一次，避免「刚上线要等 24 小时才生效」
	safeRunLoginHistoryCleanup(retentionDays)

	for {
		select {
		case <-ticker.C:
			safeRunLoginHistoryCleanup(retentionDays)
		case <-stop:
			return
		}
	}
}

// safeRunLoginHistoryCleanup 包一层 recover。
//
// 后台 goroutine 里的 panic 不会被 gin 的 Recovery 中间件接住，
// 会直接终止整个进程 —— 清理任务属于「尽力而为」的辅助功能，
// 绝不能因为它让业务服务挂掉。
func safeRunLoginHistoryCleanup(retentionDays int) {
	defer func() {
		if r := recover(); r != nil {
			facade.Errorf("登录历史清理发生 panic（已忽略，不影响业务）: %v", r)
		}
	}()
	runLoginHistoryCleanup(retentionDays)
}

// runLoginHistoryCleanup 删除超过保留期的登录历史。
// 失败只记日志，不中断后台任务（下次再试）。
func runLoginHistoryCleanup(retentionDays int) {
	// 数据库尚未就绪时直接跳过：单元测试、以及宿主还没 bootstrap 完 DB 的场景
	// 都会走到这里，此时调 DAO 会因全局 DB 为 nil 而崩。
	if framedb.GetDB() == nil {
		facade.Warnf("数据库未初始化，跳过登录历史清理")
		return
	}

	// 后台任务没有请求上下文，用 Background（DAO 内部只用它取 DB 句柄）
	ctx := context.Background()
	if err := dao.GetLoginHistoryDao().DeleteOldRecords(ctx, retentionDays); err != nil {
		facade.Errorf("清理登录历史失败（保留 %d 天）: %v", retentionDays, err)
		return
	}
	facade.Infof("登录历史清理完成（保留 %d 天）", retentionDays)
}
