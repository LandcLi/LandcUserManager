package service

import (
	"context"

	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/model"
)

var (
	userService UserService
)

func SetUserService(impl UserService) {
	userService = impl
}

func GetUserService() UserService {
	return userService
}

// TODO 扫码登录、生物登录
// UserService 认证服务
type UserService interface {
	// Login 密码验证
	Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error)

	// Register 用户注册
	Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error)

	// LoginRecord 登录审计，记录登录历史和更新信任设备，由SSO回调
	LoginRecord(ctx context.Context, req *model.LoginRecordRequest)

	// SendVerifyCode 发送验证码
	SendVerifyCode(req *model.SendVerifyCodeRequest) error

	// GetUserInfo 获取用户信息
	GetUserInfo(userID uint) (*model.UserResponse, error)

	// UpdateUserInfo 更新用户信息
	UpdateUserInfo(userID uint, nickname string, phone string, avatar string) (*model.UserResponse, error)

	// ChangePassword 修改密码
	ChangePassword(userID uint, oldPassword string, newPassword string) error

	// ResetPassword 重置密码
	ResetPassword(email string, verifyCode string, newPassword string) error

	// SendResetPasswordCode 发送重置密码验证码
	SendResetPasswordCode(email string) error
}
