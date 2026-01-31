package sdk

import (
	"context"

	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
)

// User 用户相关接口
type User struct {
	client *Client
}

// NewUser 创建用户相关接口实例
func NewUser(client *Client) *User {
	return &User{client: client}
}

// Login 登录
func (u *User) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	resp := &v1.LoginResponse{}
	err := u.client.post(ctx, "/api/user/login", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Register 注册
func (u *User) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	resp := &v1.RegisterResponse{}
	err := u.client.post(ctx, "/api/user/register", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// LoginRecord 登录记录
func (u *User) LoginRecord(ctx context.Context, req *v1.LoginRecordRequest) (*v1.LoginRecordResponse, error) {
	resp := &v1.LoginRecordResponse{}
	err := u.client.post(ctx, "/api/user/login-record", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUserInfo 获取用户信息
func (u *User) GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.GetUserInfoResponse, error) {
	resp := &v1.GetUserInfoResponse{}
	err := u.client.post(ctx, "/api/user/get-user-info", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateUserInfo 更新用户信息
func (u *User) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoResponse, error) {
	resp := &v1.UpdateUserInfoResponse{}
	err := u.client.post(ctx, "/api/user/update-user-info", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ChangePassword 修改密码
func (u *User) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error) {
	resp := &v1.ChangePasswordResponse{}
	err := u.client.post(ctx, "/api/user/change-password", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ResetPassword 重置密码
func (u *User) ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) (*v1.ResetPasswordResponse, error) {
	resp := &v1.ResetPasswordResponse{}
	err := u.client.post(ctx, "/api/user/reset-password", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
