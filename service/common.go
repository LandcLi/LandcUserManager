package service

import (
	"context"

	"github.com/LandcLi/LandcUserManager/model"
)

var (
	commonService CommonService
)

func RegisterCommonService(impl CommonService) {
	commonService = impl
}

func GetCommonService() CommonService {
	return commonService
}

// CommonService 通用服务
type CommonService interface {
	// SendVerifyCode 发送验证码
	SendVerifyCode(ctx context.Context, req *model.SendVerifyCodeRequest) error

	// CheckVerifyCode 验证验证码
	CheckVerifyCode(ctx context.Context, identifier string, verifyType string, verifyCode string, action string) error
}
