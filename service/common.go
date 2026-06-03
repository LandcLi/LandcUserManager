package service

import (
	"context"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// CommonService defines common operations interface.
type CommonService interface {
	// SendVerifyCode sends verification code
	SendVerifyCode(ctx context.Context, req *model.SendVerifyCodeRequest) error

	// CheckVerifyCode validates verification code
	CheckVerifyCode(ctx context.Context, identifier string, verifyType string, verifyCode string, action string) error
}

func RegisterCommonService(impl CommonService) {
	di.Provide[CommonService]("common.service", impl)
}

func GetCommonService() CommonService {
	return di.Require[CommonService]("common.service")
}
