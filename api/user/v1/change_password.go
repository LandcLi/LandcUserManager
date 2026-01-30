package v1

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	UserID      uint   `json:"userId" binding:"required"`
	OldPassword  string `json:"oldPassword" binding:"required,min=6,max=50"`
	NewPassword  string `json:"newPassword" binding:"required,min=6,max=50"`
}

// ChangePasswordResponse 修改密码响应
type ChangePasswordResponse struct {
	Success bool `json:"success"`
}
