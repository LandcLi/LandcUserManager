package user

import (
	"context"

	userApi "github.com/LandcLi/LandcUserManager/api/user"
	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/model"
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

func (uc *userController) LoginRecord(ctx context.Context, req *v1.LoginRecordRequest) (*v1.LoginRecordResponse, error) {
	serviceReq := &model.LoginRecordRequest{
		UserID:            req.UserID,
		VerifyType:        req.VerifyType,
		Remark:            req.Remark,
		IP:                req.IP,
		DeviceFingerprint: req.DeviceFingerprint,
		UserAgent:         req.UserAgent,
		Device:            req.Device,
		Status:            req.Status,
	}
	uc.UserService.LoginRecord(ctx, serviceReq)
	return &v1.LoginRecordResponse{
		Success: true,
	}, nil
}

func (uc *userController) GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.GetUserInfoResponse, error) {
	user, err := uc.UserService.GetUserInfo(req.UserID)
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
	user, err := uc.UserService.UpdateUserInfo(req.UserID, req.Nickname, req.Phone, req.Avatar)
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
	err := uc.UserService.ChangePassword(req.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &v1.ChangePasswordResponse{
		Success: true,
	}, nil
}

func (uc *userController) ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) (*v1.ResetPasswordResponse, error) {
	err := uc.UserService.ResetPassword(req.Email, req.VerifyCode, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &v1.ResetPasswordResponse{
		Success: true,
	}, nil
}
