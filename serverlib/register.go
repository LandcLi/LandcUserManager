package serverlib

import (
	"time"

	"fmt"

	ginmw "github.com/LandcLi/landc-go/api/middleware/gin"
	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	"github.com/LandcLi/landc-go/frame/pkg/web"
	"github.com/LandcLi/landc-go/tools/security"
	"github.com/gin-gonic/gin"

	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
	"github.com/LandcLi/LandcUserManager/internal/middleware"
	"github.com/LandcLi/LandcUserManager/service"
	outboundclient "github.com/LandcLi/LandcUserManager/utility/client"
	"github.com/LandcLi/LandcUserManager/utility/config"

	_ "github.com/LandcLi/LandcUserManager/internal" // 触发默认实现注册链（di.Provide）
)

// options 为库入口配置。
type options struct {
	authEnabled bool
	webOpts     []web.RegisterOption
}

// Option 为库入口配置项。
type Option func(*options)

// WithAuth 是否装配与独立部署一致的全局鉴权中间件（默认 true）。
func WithAuth(enabled bool) Option {
	return func(o *options) { o.authEnabled = enabled }
}

// WithWebOptions 透传 landc-go 注册选项（前缀 / 方法覆盖 / 方法级与控制器级中间件）。
func WithWebOptions(opts ...web.RegisterOption) Option {
	return func(o *options) { o.webOpts = append(o.webOpts, opts...) }
}

// RegisterToRouter 装配 user 服务的默认实现并注册到 server。
// 只装配、不接管：不创建 engine、不调用 Run；调用方可继续追加自有路由。
// 前置条件：宿主进程已完成 landc-go 框架 bootstrap（配置/DB/缓存/JWT）。
// Install 只做「装配」：确保表结构、初始化第三方登录适配器，
// 把本系统的控制器实现留在 DI 容器中，**不**挂中间件、**不**注册 HTTP 路由。
//
// 适用场景：宿主服务（如 LandcSSO）想通过 `user.UserGateway.Get()` 在同一进程内
// 直接调用用户服务，而不希望 LUM 的鉴权中间件与路由影响宿主的 URL 空间与中间件链。
// 需要 LUM 同时对外提供 HTTP 接口时改用 RegisterToRouter。
//
// 幂等，可重复调用。
func Install() error {
	// 防御性检查：框架配置未初始化时返回友好错误，而非 panic
	if frameconfig.GetConfig() == nil {
		return fmt.Errorf("serverlib: framework not bootstrapped; initialize global config first")
	}

	// 自身表结构：受 database.auto_migrate 控制（SQLite/开发环境开箱建表，
	// MySQL 生产环境交由 sqls/init.sql 或版本化迁移），幂等
	if err := EnsureSchema(); err != nil {
		return err
	}

	// 第三方登录适配器（未配置 third_party 时为 no-op）
	cfg := config.GetConfig()
	if cfg != nil {
		if err := service.InitWithConfig(&cfg.ThirdParty); err != nil {
			return fmt.Errorf("serverlib: init third-party adapters: %w", err)
		}
	}
	return nil
}

// RegisterToRouter 装配默认实现并把控制器注册到 server。
// 只装配、不接管：不创建 engine、不调用 Run；调用方可继续追加自有路由。
// 前置条件：宿主进程已完成 landc-go 框架 bootstrap（配置/DB/缓存/JWT）。
func RegisterToRouter(s *web.Server, opts ...Option) error {
	o := &options{authEnabled: true}
	for _, opt := range opts {
		opt(o)
	}

	if err := Install(); err != nil {
		return err
	}

	// gin 运行模式：ServerConfig.GinMode 此前只声明字段、从未赋值，等于配了不生效。
	// （debug 模式会打印全部路由并在日志里输出调试信息，生产应切 release）
	if cfg := config.TryGetConfig(); cfg != nil {
		if cfg.Server.GinMode != "" {
			gin.SetMode(cfg.Server.GinMode)
		}
		// 出站 HTTP 超时：http_client.timeout 此前从未被读取，
		// adapter 各自 new client（GitHub 那个甚至没有超时）。
		if cfg.HTTPClient.Timeout > 0 {
			outboundclient.SetHTTPTimeout(time.Duration(cfg.HTTPClient.Timeout) * time.Second)
		}
		// 安全响应头：配置里一直有开关，但代码从未实现，响应里一个头都没有。
		// 框架提供了 SecurityHeaders 中间件，这里按配置挂上。
		if cfg.Server.SecurityHeaders.Enabled {
			s.Engine().Use(ginmw.SecurityHeaders(toHeadersConfig(cfg.Server.SecurityHeaders)))
		}
	}

	// 全局鉴权中间件，必须在 RegisterHandler 之前挂载（与独立部署一致）
	if o.authEnabled {
		s.Engine().Use(middleware.Auth())
	}

	// 控制器清单：common + user
	if err := s.RegisterHandler(common.GetCommonController(), o.webOpts...); err != nil {
		return err
	}
	if err := s.RegisterHandler(user.GetUserController(), o.webOpts...); err != nil {
		return err
	}

	// 登录历史清理：DAO 早就有 DeleteOldRecords，但从未被调度过，
	// login_histories 只写不删会无限增长。
	if cfg := config.TryGetConfig(); cfg != nil {
		startLoginHistoryCleanup(cfg.LoginHistory.RetentionDays)
	}

	return nil
}

// toHeadersConfig 把本系统配置转成框架的安全头配置。
//
// 未显式设置的项取安全默认：FrameDeny / nosniff 开，HSTS 关
// （内网常见 HTTP 部署，发 HSTS 会把浏览器强制升级到 HTTPS，反而打不开）。
func toHeadersConfig(c config.SecurityHeadersConfig) security.HeadersConfig {
	out := security.HeadersConfig{
		HSTSMaxAge:         c.HSTSMaxAge,
		FrameDeny:          true,
		ContentTypeNoSniff: true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}
	if c.FrameDeny != nil {
		out.FrameDeny = *c.FrameDeny
	}
	if c.ContentTypeNoSniff != nil {
		out.ContentTypeNoSniff = *c.ContentTypeNoSniff
	}
	if c.ReferrerPolicy != "" {
		out.ReferrerPolicy = c.ReferrerPolicy
	}
	return out
}
