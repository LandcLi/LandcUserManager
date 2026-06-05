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
	err := server.RegisterHandler(common.GetCommonController())
	if err != nil {
		return err
	}
	err = server.RegisterHandler(user.GetUserController())
	if err != nil {
		return err
	}
	err = server.RunWithBootstrap()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := RunMain()
	if err != nil {
		log.Fatalf("RunMain failed: %v", err)
	}
}
