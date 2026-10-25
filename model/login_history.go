package model

import (
	"time"
)

// LoginHistory 登录历史模型
type LoginHistory struct {
	ID uint `json:"id" gorm:"primaryKey"`
	// UserID 为 0 表示「标识符不存在」的失败尝试（此时查不到用户，
	// 但这类记录正是撞库/枚举的最直接证据，必须留痕）。
	UserID uint `json:"userId" gorm:"index;not null"`
	// Identifier 本次登录使用的标识符（邮箱/手机号/用户名，已小写归一化）。
	// 失败记录靠它定位「哪个账号正在被爆破」，仅有 user_id 时无从查起。
	Identifier        string    `json:"identifier" gorm:"size:100;index"`
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
