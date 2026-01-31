package service

import (
	"context"

	"github.com/LandcLi/LandcUserManager/model"
)

var (
	commonServiceImpl commonService
)

func RegisterCommonService(impl commonService) {
	commonServiceImpl = impl
}

func GetCommonService() commonService {
	return commonServiceImpl
}

// commonService 通用服务
type commonService interface {
	// SendVerifyCode 发送验证码
	SendVerifyCode(ctx context.Context, req *model.SendVerifyCodeRequest) error

	// CheckVerifyCode 验证验证码
	CheckVerifyCode(ctx context.Context, identifier string, verifyType string, verifyCode string, action string) error
}
