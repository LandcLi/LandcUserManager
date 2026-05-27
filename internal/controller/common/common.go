package common

import (
	"context"

	commonApi "github.com/LandcLi/LandcUserManager/api/common"
	v1 "github.com/LandcLi/LandcUserManager/api/common/v1"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/landc-go/mvc/pkg/di"
)

func init() {
	di.Provide[commonApi.CommonController]("common.controller", NewCommonController())
}

type commonController struct {
}

func NewCommonController() *commonController {
	return &commonController{}
}

func (cc *commonController) SendVerifyCode(ctx context.Context, req *v1.SendVerifyCodeRequest) (*v1.SendVerifyCodeResponse, error) {
	serviceReq := &model.SendVerifyCodeRequest{
		Identifier: req.Identifier,
		VerifyType: req.VerifyType,
		Action:     req.Action,
	}
	err := service.GetCommonService().SendVerifyCode(ctx, serviceReq)
	if err != nil {
		return nil, err
	}
	return &v1.SendVerifyCodeResponse{
		Success: true,
	}, nil
}
