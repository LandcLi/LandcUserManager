package model

import (
	"time"
)

// Token 令牌模型
type Token struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"userId" gorm:"index;not null"`
	AccessToken  string    `json:"accessToken" gorm:"size:255;not null"`
	RefreshToken string    `json:"refreshToken" gorm:"size:255;not null"`
	ExpiresAt    time.Time `json:"expiresAt"`
	RefreshAt    time.Time `json:"refreshAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (Token) TableName() string {
	return "tokens"
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresIn    int64        `json:"expiresIn"`
	TokenType    string       `json:"tokenType"`
	User         *UserResponse `json:"user"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
