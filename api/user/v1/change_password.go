package v1

import "github.com/LandcLi/landc-go/frame/pkg/meta"

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	meta.Meta `path:"/api/user/change-password" method:"POST"`
	UserID      uint   `json:"userId" binding:"required"`
	OldPassword  string `json:"oldPassword" binding:"required,min=6,max=50"`
	NewPassword  string `json:"newPassword" binding:"required,min=6,max=50"`
}

// ChangePasswordResponse 修改密码响应
type ChangePasswordResponse struct {
	Success bool `json:"success"`
}
