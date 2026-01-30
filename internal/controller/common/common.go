package common

import (
	"context"

	commonApi "github.com/LandcLi/LandcUserManager/api/common"
	v1 "github.com/LandcLi/LandcUserManager/api/common/v1"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/service"
)

func init() {
	if commonApi.GetCommonController() == nil {
		commonApi.RegisterCommonController(NewCommonController())
	}
}

type commonController struct {
	CommonService service.CommonService
}

func NewCommonController() *commonController {
	return &commonController{
		CommonService: service.GetCommonService(),
	}
}

func (cc *commonController) SendVerifyCode(ctx context.Context, req *v1.SendVerifyCodeRequest) (*v1.SendVerifyCodeResponse, error) {
	serviceReq := &model.SendVerifyCodeRequest{
		Identifier: req.Identifier,
		VerifyType: req.VerifyType,
		Action:     req.Action,
	}
	err := cc.CommonService.SendVerifyCode(ctx, serviceReq)
	if err != nil {
		return nil, err
	}
	return &v1.SendVerifyCodeResponse{
		Success: true,
	}, nil
}
