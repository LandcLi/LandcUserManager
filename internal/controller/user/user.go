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
