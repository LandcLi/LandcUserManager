package user

import (
	"context"

	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// UserController defines the user management interface.
// Both local implementation and remote SDK implement this interface.
type UserController interface {
	Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error)
	Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error)
	LoginRecord(ctx context.Context, req *v1.LoginRecordRequest) (*v1.LoginRecordResponse, error)
	GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.GetUserInfoResponse, error)
	UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoResponse, error)
	ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error)
	ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) (*v1.ResetPasswordResponse, error)
}

// UserRoutes maps controller methods to HTTP endpoints.
var UserRoutes = di.Routes{
	"Login":          {Method: "POST", Path: "/api/user/login"},
	"Register":       {Method: "POST", Path: "/api/user/register"},
	"LoginRecord":    {Method: "POST", Path: "/api/user/login-record"},
	"GetUserInfo":    {Method: "POST", Path: "/api/user/get-user-info"},
	"UpdateUserInfo": {Method: "POST", Path: "/api/user/update-user-info"},
	"ChangePassword": {Method: "POST", Path: "/api/user/change-password"},
	"ResetPassword":  {Method: "POST", Path: "/api/user/reset-password"},
}

// UserGateway is the service gateway for user management.
// It provides automatic Gin handler generation and remote proxy support.
//
// Local mode (in init() of controller implementation):
//
//	user.UserGateway.Provide(localImpl)
//
// Remote mode (in main.go):
//
//	user.UserGateway.Override(sdk.NewRemoteUserController(url))
var UserGateway = di.NewGateway[UserController]("user.controller", UserRoutes)

// GetUserController returns the registered UserController implementation.
// Kept for backward compatibility with existing code (e.g., SSO's auth controller).
func GetUserController() UserController {
	return UserGateway.Get()
}

// RegisterUserController registers a UserController implementation.
// Kept for backward compatibility.
func RegisterUserController(impl UserController) {
	UserGateway.Provide(impl)
}
