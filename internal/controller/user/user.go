package user

import (
	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/frame/pkg/web"

	userApi "github.com/LandcLi/LandcUserManager/api/user"
	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
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
func (uc *userController) Login(ctx *web.LandcContext, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	return service.GetUserService().Login(ctx.GetGinContext(), req)
}

func (uc *userController) Register(ctx *web.LandcContext, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	return service.GetUserService().Register(ctx.GetGinContext(), req)
}

func (uc *userController) LoginRecord(ctx *web.LandcContext, req *v1.LoginRecordRequest) (*v1.LoginRecordResponse, error) {
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
	service.GetUserService().LoginRecord(ctx.GetGinContext(), serviceReq)
	return &v1.LoginRecordResponse{
		Success: true,
	}, nil
}

func (uc *userController) GetUserInfo(ctx *web.LandcContext, req *v1.GetUserInfoRequest) (*v1.GetUserInfoResponse, error) {
	user, err := service.GetUserService().GetUserInfo(req.UserID)
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

func (uc *userController) UpdateUserInfo(ctx *web.LandcContext, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoResponse, error) {
	user, err := service.GetUserService().UpdateUserInfo(req.UserID, req.Nickname, req.Phone, req.Avatar)
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

func (uc *userController) ChangePassword(ctx *web.LandcContext, req *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error) {
	err := service.GetUserService().ChangePassword(req.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &v1.ChangePasswordResponse{
		Success: true,
	}, nil
}

func (uc *userController) ResetPassword(ctx *web.LandcContext, req *v1.ResetPasswordRequest) (*v1.ResetPasswordResponse, error) {
	err := service.GetUserService().ResetPassword(req.Email, req.VerifyCode, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &v1.ResetPasswordResponse{
		Success: true,
	}, nil
}
