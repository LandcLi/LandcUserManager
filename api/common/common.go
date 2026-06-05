package common

import (
	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/frame/pkg/web"

	v1 "github.com/LandcLi/LandcUserManager/api/common/v1"
)

// CommonController defines the common operations interface.
type CommonController interface {
	SendVerifyCode(ctx *web.LandcContext, req *v1.SendVerifyCodeRequest) (*v1.SendVerifyCodeResponse, error)
}

// CommonGateway is the service gateway for common operations.
var CommonGateway = di.NewGateway[CommonController]("common.controller")

// GetCommonController returns the registered CommonController implementation.
func GetCommonController() CommonController {
	return CommonGateway.Get()
}

// RegisterCommonController registers a CommonController implementation.
func RegisterCommonController(impl CommonController) {
	CommonGateway.Provide(impl)
}
