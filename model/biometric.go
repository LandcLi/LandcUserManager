package model

type Biometric struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	UserID        uint   `json:"userId" gorm:"index"`
	BiometricType string `json:"biometricType" gorm:"size:20"`
	BiometricData string `json:"biometricData" gorm:"size:255"`
}

// TableName 指定表名
func (Biometric) TableName() string {
	return "biometrics"
}

// BiometricLoginRequest 生物识别登录请求
type BiometricLoginRequest struct {
	UserID            uint   `json:"userId" binding:"required"`
	DeviceFingerprint string `json:"deviceFingerprint" binding:"required"`
	BiometricData     string `json:"biometricData" binding:"required"`
	BiometricType     string `json:"biometricType" binding:"required,oneof=fingerprint face iris"`
}
