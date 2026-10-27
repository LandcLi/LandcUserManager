package client

import (
	"net/http"
	"sync"
	"time"
)

// defaultHTTPTimeout 出站 HTTP 调用的默认超时。
// 与 config.example.yml 的 http_client.timeout 保持一致。
const defaultHTTPTimeout = 10 * time.Second

var (
	httpClientOnce sync.Once
	sharedHTTPClient *http.Client

	// httpTimeout 由 SetHTTPTimeout 注入。
	// 不在本包直接读 config 是为了避免循环依赖
	// （utility/config 已经依赖 utility/client 来构造邮件客户端）。
	httpTimeout = defaultHTTPTimeout
)

// SetHTTPTimeout 设置出站请求超时，应在配置加载完成后调用一次。
//
// 调用时机早于首次 GetHTTPClient 才会生效；晚于则本次进程沿用旧值。
func SetHTTPTimeout(d time.Duration) {
	if d > 0 {
		httpTimeout = d
	}
}

// GetHTTPClient 返回带超时的共享 HTTP 客户端，供第三方登录等出站调用使用。
//
// 为什么需要这个函数：此前各 adapter 自己 new http.Client ——
//   - 钉钉写死 10 秒；
//   - GitHub 用的是 &http.Client{}，**完全没有超时**，下游不响应时会永久挂起，
//     连接池被逐个占满后整个服务失去对外能力；
//   - 配置里的 http_client.timeout 则从未被读取，配了也不生效。
//
// 共享同一个 client 还能复用 TCP 连接与连接池，比每次请求新建高效。
func GetHTTPClient() *http.Client {
	httpClientOnce.Do(func() {
		sharedHTTPClient = &http.Client{
			Timeout: httpTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	})
	return sharedHTTPClient
}
