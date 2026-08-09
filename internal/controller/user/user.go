package user

import (
	"context"

	"github.com/LandcLi/landc-go/frame/pkg/di"

	userApi "github.com/LandcLi/LandcUserManager/api/user"
	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/internal/middleware"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/service"
)

func init() {
	// Register local controller implementation via Gateway
	di.Provide[userApi.UserController]("user.controller", NewUserController())
}

type userController struct {
}

func NewUserController() *userController {
	return &userController{}
}

// currentUserID 从 request context 读取 Auth 中间件注入的 UserID（JWT claims）。
// 业务一律使用该方法获取当前登录用户，禁止信任前端传参（防 IDOR 越权）。
func currentUserID(ctx context.Context) uint {
	id, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return 0
	}
	return id
}

func (uc *userController) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	return service.GetUserService().Login(ctx, req)
}

func (uc *userController) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	return service.GetUserService().Register(ctx, req)
}

func (uc *userController) LoginRecord(ctx context.Context, req *v1.LoginRecordRequest) (*v1.LoginRecordResponse, error) {
	serviceReq := &model.LoginRecordRequest{
		UserID:            currentUserID(ctx),
		VerifyType:        req.VerifyType,
		Remark:            req.Remark,
		IP:                req.IP,
		DeviceFingerprint: req.DeviceFingerprint,
		UserAgent:         req.UserAgent,
		Device:            req.Device,
		Status:            req.Status,
	}
	service.GetUserService().LoginRecord(ctx, serviceReq)
	return &v1.LoginRecordResponse{
		Success: true,
	}, nil
}

func (uc *userController) GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.GetUserInfoResponse, error) {
	user, err := service.GetUserService().GetUserInfo(ctx, currentUserID(ctx))
	if err != nil {
		return nil, err
	}
	return &v1.GetUserInfoResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Phone:     user.Phone,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (uc *userController) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoResponse, error) {
	user, err := service.GetUserService().UpdateUserInfo(ctx, currentUserID(ctx), req.Nickname, req.Phone, req.Avatar)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateUserInfoResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Phone:     user.Phone,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (uc *userController) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error) {
	err := service.GetUserService().ChangePassword(ctx, currentUserID(ctx), req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &v1.ChangePasswordResponse{
		Success: true,
	}, nil
}

func (uc *userController) ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) (*v1.ResetPasswordResponse, error) {
	err := service.GetUserService().ResetPassword(ctx, req.Email, req.VerifyCode, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &v1.ResetPasswordResponse{
		Success: true,
	}, nil
}
