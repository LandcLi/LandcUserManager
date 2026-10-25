package serverlib

import (
	"fmt"

	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	"github.com/LandcLi/landc-go/frame/pkg/web"

	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
	"github.com/LandcLi/LandcUserManager/internal/middleware"
	"github.com/LandcLi/LandcUserManager/service"
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
func RegisterToRouter(s *web.Server, opts ...Option) error {
	o := &options{authEnabled: true}
	for _, opt := range opts {
		opt(o)
	}

	// 防御性检查：框架配置未初始化时返回友好错误，而非 panic
	if frameconfig.GetConfig() == nil {
		return fmt.Errorf("serverlib: framework not bootstrapped; initialize global config first")
	}

	// 自身表结构：受 database.auto_migrate 控制（SQLite/开发环境开箱建表，
	// MySQL 生产环境交由 sqls/init.sql 或版本化迁移），幂等
	if err := EnsureSchema(); err != nil {
		return err
	}

	// 补全（相对 gen lib 骨架）：全局鉴权中间件，必须在 RegisterHandler 之前挂载（与独立部署一致）
	if o.authEnabled {
		s.Engine().Use(middleware.Auth())
	}

	// 补全（相对 gen lib 骨架）：第三方登录适配器初始化
	if err := service.InitWithConfig(&config.GetConfig().ThirdParty); err != nil {
		return fmt.Errorf("serverlib: init third-party adapters: %w", err)
	}

	// 控制器清单（gen lib 提示补全）：common + user
	if err := s.RegisterHandler(common.GetCommonController(), o.webOpts...); err != nil {
		return err
	}
	return s.RegisterHandler(user.GetUserController(), o.webOpts...)
}
