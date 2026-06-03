package service

import (
	"context"

	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// UserService defines the user business logic interface.
// Users can implement this interface and register via di.Override to customize behavior.
type UserService interface {
	// Login validates credentials
	Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error)

	// Register creates a new user
	Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error)

	// LoginRecord records login history and updates trusted devices
	LoginRecord(ctx context.Context, req *model.LoginRecordRequest)

	// GetUserInfo retrieves user information
	GetUserInfo(userID uint) (*model.User, error)

	// UpdateUserInfo updates user information
	UpdateUserInfo(userID uint, nickname string, phone string, avatar string) (*model.User, error)

	// ChangePassword changes user password
	ChangePassword(userID uint, oldPassword string, newPassword string) error

	// ResetPassword resets user password
	ResetPassword(email string, verifyCode string, newPassword string) error

	// SendResetPasswordCode sends reset password verification code
	SendResetPasswordCode(email string) error
}

func RegisterUserService(impl UserService) {
	di.Provide[UserService]("user.service", impl)
}

func GetUserService() UserService {
	return di.Require[UserService]("user.service")
}
