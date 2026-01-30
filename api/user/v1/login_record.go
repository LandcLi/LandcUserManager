package v1

// LoginRecordRequest 登录记录请求
type LoginRecordRequest struct {
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
