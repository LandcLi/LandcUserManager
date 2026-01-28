package model

import (
	"time"
)

// TrustedDevice 信任设备模型
type TrustedDevice struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"userId" gorm:"index;not null"`
	DeviceFingerprint string    `json:"deviceFingerprint" gorm:"size:255;not null"`
	DeviceName        string    `json:"deviceName" gorm:"size:100"`
	LastLoginIP       string    `json:"lastLoginIp" gorm:"size:50"`
	FirstLoginTime    time.Time `json:"firstLoginTime"`
	LastLoginTime     time.Time `json:"lastLoginTime"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	LastLoginLocation string    `json:"lastLoginLocation" gorm:"size:100"`
}

// TableName 指定表名
func (TrustedDevice) TableName() string {
	return "trusted_devices"
}
