package common

import (
	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/frame/pkg/web"

	commonApi "github.com/LandcLi/LandcUserManager/api/common"
	v1 "github.com/LandcLi/LandcUserManager/api/common/v1"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/service"
)

func init() {
	di.Provide[commonApi.CommonController]("common.controller", NewCommonController())
}

type commonController struct {
}

func NewCommonController() *commonController {
	return &commonController{}
}

func (cc *commonController) SendVerifyCode(ctx *web.LandcContext, req *v1.SendVerifyCodeRequest) (*v1.SendVerifyCodeResponse, error) {
	serviceReq := &model.SendVerifyCodeRequest{
		Identifier: req.Identifier,
		VerifyType: req.VerifyType,
		Action:     req.Action,
	}
	err := service.GetCommonService().SendVerifyCode(ctx.GetGinContext(), serviceReq)
	if err != nil {
		return nil, err
	}
	return &v1.SendVerifyCodeResponse{
		Success: true,
	}, nil
}
