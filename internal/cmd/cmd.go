package cmd

import (
	"context"

	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
	"github.com/LandcLi/landc-go/frame/pkg/cmd"
	"github.com/LandcLi/landc-go/frame/pkg/web"
)

// Main is the main application command.
var Main = cmd.NewCommand("main", "start HTTP server", func(ctx context.Context, parser *cmd.Parser) error {
	server := web.NewServer(nil)
	if err := server.RegisterHandler(common.GetCommonController()); err != nil {
		return err
	}
	if err := server.RegisterHandler(user.GetUserController()); err != nil {
		return err
	}
	return server.RunWithContext(ctx)
})
