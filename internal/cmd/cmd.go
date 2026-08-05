package cmd

import (
	"context"

	"github.com/LandcLi/LandcUserManager/serverlib"
	"github.com/LandcLi/landc-go/frame/pkg/cmd"
	"github.com/LandcLi/landc-go/frame/pkg/web"
)

// Main is the main application command.
// 独立部署入口：通过 serverlib 库模式入口装配默认实现（挂载鉴权中间件、
// 初始化第三方登录适配器、注册 common 与 user 控制器），行为与直接装配一致。
var Main = cmd.NewCommand("main", "start HTTP server", func(ctx context.Context, parser *cmd.Parser) error {
	server := web.New()
	if err := serverlib.RegisterToRouter(server); err != nil {
		return err
	}
	return server.RunWithContext(ctx)
})
