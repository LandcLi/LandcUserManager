package cmd

import (
	"context"

	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
	"github.com/LandcLi/LandcUserManager/internal/middleware"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/LandcLi/landc-go/frame/pkg/cmd"
	"github.com/LandcLi/landc-go/frame/pkg/web"
)

// Main is the main application command.
var Main = cmd.NewCommand("main", "start HTTP server", func(ctx context.Context, parser *cmd.Parser) error {
	server := web.New()

	// 全局鉴权中间件：公开路径放行，其余路径校验 Bearer Token
	// 注意：必须在 RegisterHandler 之前挂载，否则已注册路由不会生效
	server.Engine().Use(middleware.Auth())

	// 初始化第三方登录适配器（根据 config.yaml 中 third_party 配置）
	if err := service.InitWithConfig(&config.GetConfig().ThirdParty); err != nil {
		return err
	}

	if err := server.RegisterHandler(common.GetCommonController()); err != nil {
		return err
	}
	if err := server.RegisterHandler(user.GetUserController()); err != nil {
		return err
	}
	return server.RunWithContext(ctx)
})
