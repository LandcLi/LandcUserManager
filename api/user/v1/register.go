package v1

import "github.com/LandcLi/landc-go/frame/pkg/meta"

// RegisterRequest 注册请求
type RegisterRequest struct {
	meta.Meta `path:"/api/user/register" method:"POST"`
	Identifier        string `json:"identifier" binding:"required,min=3,max=50"`
	Password          string `json:"password" binding:"required,min=6,max=50"`
	AuthCode          string `json:"verifyCode" binding:"required"`
	AuthType          string `json:"authType" binding:"required,oneof=PASSWORD VERIFY_CODE"`
	VerifyType        string `json:"verifyType" binding:"required,oneof=EMAIL PHONE USERNAME"`
	IP                string `json:"ip" binding:"required,ip"`
	DeviceFingerprint string `json:"deviceFingerprint" binding:"required"`
	UserAgent         string `json:"userAgent" binding:"required"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	UserID       uint   `json:"userId"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}
