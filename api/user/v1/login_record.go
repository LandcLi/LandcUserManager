package v1

import "github.com/LandcLi/landc-go/frame/pkg/meta"

// LoginRecordRequest 登录记录请求
type LoginRecordRequest struct {
	meta.Meta `path:"/api/user/login-record" method:"POST"`
	UserID            uint   `json:"userId" binding:"required"`
	VerifyType        string `json:"verifyType" binding:"required,oneof=EMAIL PHONE USERNAME"`
	Remark            string `json:"remark" binding:"required,min=3,max=255"`
	IP                string `json:"ip" binding:"required,ip"`
	DeviceFingerprint string `json:"deviceFingerprint" binding:"required"`
	UserAgent         string `json:"userAgent" binding:"required"`
	Device            string `json:"device"`
	Status            int    `json:"status"`
}

// LoginRecordResponse 登录记录响应
type LoginRecordResponse struct {
	Success bool `json:"success"`
}
