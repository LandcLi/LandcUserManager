package user

import (
	"context"

	userApi "github.com/LandcLi/LandcUserManager/api/user"
	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/service"
)

func init() {
	if userApi.GetUserController() == nil {
		userApi.RegisterUserController(NewUserController())
	}
}

type userController struct {
	UserService service.UserService
}

func NewUserController() *userController {
	return &userController{
		UserService: service.GetUserService(),
	}
}
func (uc *userController) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	return uc.UserService.Login(ctx, req)
}

func (uc *userController) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	return uc.UserService.Register(ctx, req)
}
