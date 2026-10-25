package model

import (
	"time"
)

// User 用户模型,username和nickname的区别:
// username: 唯一标识用户, 用于登录, 限制修改，仅能为字母、数字、下划线，至少8个字符
// nickname: 可修改的用户显示名称, 可以与username不同
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"uniqueIndex;size:50;not null"`
	// Email 可空：用户名注册、手机号注册、第三方登录都可能没有邮箱。
	// 早期这里是 NOT NULL + UNIQUE，而空串在 SQL 里是确定值，导致「不填邮箱的
	// 用户」全表只能存在一个 —— 第二个用户名注册就 UNIQUE constraint failed。
	// 现在改为普通索引，非空邮箱的唯一性由应用层在注册/改绑时校验。
	Email     string    `json:"email" gorm:"index;size:100"`
	Password  string    `json:"-" gorm:"size:100;not null"`
	Nickname  string    `json:"nickname" gorm:"size:50"`
	Avatar    string    `json:"avatar" gorm:"size:255"`
	Phone     string    `json:"phone" gorm:"size:20"`
	Status    int       `json:"status" gorm:"default:1"` // 1: 活跃, 0: 禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

type UserDto struct {
	ID            uint      `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Nickname      string    `json:"nickname"`
	Avatar        string    `json:"avatar"`
	Phone         string    `json:"phone"`
	Status        int       `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ThirdPartyIDs []uint    `json:"thirdPartyIDs"`
	BiometricIDs  []uint    `json:"biometricIDs"`
}

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

// SendVerifyCodeRequest 验证码请求
type SendVerifyCodeRequest struct {
	Identifier string `json:"identifier" binding:"required,min=3,max=50"`
	VerifyType string `json:"verifyType" binding:"required,oneof=EMAIL PHONE"`
	Action     string `json:"action" binding:"required,oneof=login register reset"`
}

// VerifyCodeLoginRequest 验证码登录请求
type VerifyCodeLoginRequest struct {
	Identifier        string `json:"identifier" binding:"required,min=3,max=50"`
	VerifyType        string `json:"verifyType" binding:"required,oneof=EMAIL PHONE"`
	VerifyCode        string `json:"verifyCode" binding:"required,len=6"`
	IP                string `json:"ip" binding:"required,ip"`
	DeviceFingerprint string `json:"deviceFingerprint" binding:"required"`
	UserAgent         string `json:"userAgent" binding:"required"`
}

// ThirdPartyLoginRequest 第三方登录请求
type ThirdPartyLoginRequest struct {
	Provider string `json:"provider" binding:"required,oneof=wechat dingtalk github"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
}

// UserRegisterRequest 注册请求
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
}
