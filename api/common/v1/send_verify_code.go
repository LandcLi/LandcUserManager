package v1

// SendVerifyCodeRequest 发送验证码请求
type SendVerifyCodeRequest struct {
	Identifier string `json:"identifier" binding:"required,min=3,max=50"`
	VerifyType string `json:"verifyType" binding:"required,oneof=EMAIL PHONE"`
	Action     string `json:"action" binding:"required,oneof=LOGIN REGISTER RESET"`
}

// SendVerifyCodeResponse 发送验证码响应
type SendVerifyCodeResponse struct {
	Success bool `json:"success"`
}
