package service

import (
	"context"

	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/model"
)

var (
	userServiceImpl userService
)

func RegisterUserService(impl userService) {
	userServiceImpl = impl
}

func GetUserService() userService {
	return userServiceImpl
}

// TODO 扫码登录、生物登录
// userService 认证服务
type userService interface {
	// Login 密码验证
	Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error)

	// Register 用户注册
	Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error)

	// LoginRecord 登录审计，记录登录历史和更新信任设备，由SSO回调
	LoginRecord(ctx context.Context, req *model.LoginRecordRequest)

	// GetUserInfo 获取用户信息
	GetUserInfo(userID uint) (*model.User, error)

	// UpdateUserInfo 更新用户信息
	UpdateUserInfo(userID uint, nickname string, phone string, avatar string) (*model.User, error)

	// ChangePassword 修改密码
	ChangePassword(userID uint, oldPassword string, newPassword string) error

	// ResetPassword 重置密码
	ResetPassword(email string, verifyCode string, newPassword string) error

	// SendResetPasswordCode 发送重置密码验证码
	SendResetPasswordCode(email string) error
}
