package v1

import "github.com/LandcLi/landc-go/frame/pkg/meta"

// LoginRecordRequest 登录记录请求
// 注意：UserID 不从前端获取，而是由 Auth 中间件从 JWT claims 中解析并注入。
type LoginRecordRequest struct {
	meta.Meta         `path:"/api/user/login-record" method:"POST"`
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
