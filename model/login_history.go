package model

import (
	"time"
)

// LoginHistory 登录历史模型
type LoginHistory struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"userId" gorm:"index;not null"`
	IP                string    `json:"ip" gorm:"size:50"`
	DeviceFingerprint string    `json:"deviceFingerprint" gorm:"size:255"`
	UserAgent         string    `json:"userAgent" gorm:"size:255"`
	LoginTime         time.Time `json:"loginTime"`
	Status            int       `json:"status"` // 1: 成功, 0: 失败
	Remark            string    `json:"remark" gorm:"size:255"`
	CreatedAt         time.Time `json:"createdAt"`
	Device            string    `json:"device" gorm:"size:100"`
	Location          string    `json:"location" gorm:"size:100"`
}

// TableName 指定表名
func (LoginHistory) TableName() string {
	return "login_histories"
}
