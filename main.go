package main

import (
	"log"

	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
	_ "github.com/LandcLi/LandcUserManager/internal"

	"github.com/LandcLi/landc-go/frame/pkg/web"
)

func RunMain() error {
	server := web.NewServer(nil)
	if err := server.RegisterHandler(common.GetCommonController()); err != nil {
		return err
	}
	if err := server.RegisterHandler(user.GetUserController()); err != nil {
		return err
	}
	return server.RunWithBootstrap()
}

func main() {
	if err := RunMain(); err != nil {
		log.Fatalf("RunMain failed: %v", err)
	}
}
