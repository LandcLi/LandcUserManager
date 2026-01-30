package v1

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	VerifyCode  string `json:"verifyCode" binding:"required,len=6"`
	NewPassword  string `json:"newPassword" binding:"required,min=6,max=50"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Success bool `json:"success"`
}
