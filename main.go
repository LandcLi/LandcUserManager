package main

import (
	"context"
	"log"

	_ "github.com/LandcLi/LandcUserManager/internal"
	"github.com/LandcLi/LandcUserManager/internal/cmd"
)

func main() {
	if err := cmd.Main.Run(context.Background()); err != nil {
		log.Fatalf("RunMain failed: %v", err)
	}
}
