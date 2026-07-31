package v1

import "github.com/LandcLi/landc-go/frame/pkg/meta"

// SendVerifyCodeRequest 发送验证码请求
	type SendVerifyCodeRequest struct {
		meta.Meta  `path:"/api/common/send-verify-code" method:"POST"`
	Identifier string `json:"identifier" binding:"required,min=3,max=50"`
	VerifyType string `json:"verifyType" binding:"required,oneof=EMAIL PHONE"`
	Action     string `json:"action" binding:"required,oneof=login register reset"`
}

// SendVerifyCodeResponse 发送验证码响应
type SendVerifyCodeResponse struct {
	Success bool `json:"success"`
}
