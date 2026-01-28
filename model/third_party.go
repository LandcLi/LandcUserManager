package model

import (
	"time"
)

// ThirdPartyUser 获取到的第三方用户信息
type ThirdPartyUser struct {
	Provider    string `json:"provider"`
	OpenID      string `json:"openId"`
	UnionID     string `json:"unionId"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Gender      string `json:"gender"`
	Country     string `json:"country"`
	Province    string `json:"province"`
	City        string `json:"city"`
	AccessToken string `json:"accessToken"`
}

// ThirdPartyUserMeta 第三方用户的最小元数据，保存在数据库中
type ThirdPartyUserMeta struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"userId" gorm:"index;not null"`
	Provider    string    `json:"provider" gorm:"size:50;not null"`
	OpenID      string    `json:"openId" gorm:"size:100;not null"`
	UnionID     string    `json:"unionId" gorm:"size:100"`
	AccessToken string    `json:"accessToken" gorm:"size:255"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (ThirdPartyUserMeta) TableName() string {
	return "third_party_user_metas"
}
