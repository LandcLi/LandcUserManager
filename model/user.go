package model

import (
	"time"
)

// User 用户模型,username和nickname的区别:
// username: 唯一标识用户, 用于登录, 限制修改，仅能为字母、数字、下划线，至少8个字符
// nickname: 可修改的用户显示名称, 可以与username不同
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
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
type LoginRequest struct {
	Identifier        string `json:"identifier" binding:"required,min=3,max=50"`
	AuthCode          string `json:"verifyCode" binding:"required,len=6"` //第三方登陆时为openid,sso系统获取code、access_token然后再调用此接口
	AccessToken       string `json:"accessToken"`
	RefreshToken      string `json:"refreshToken"`
	VerifyType        string `json:"verifyType" binding:"required,oneof=EMAIL PHONE USERNAME WECHAT GITHUB ALIPAY DINGTALK"`
	AuthType          string `json:"authType" binding:"required,oneof=PASSWORD VERIFY_CODE THIRD_PARTY"`
	IP                string `json:"ip" binding:"required,ip"`
	DeviceFingerprint string `json:"deviceFingerprint" binding:"required"`
	UserAgent         string `json:"userAgent" binding:"required"`
}

// LoginResponse 密码验证响应
type LoginResponse struct {
	UserID            uint `json:"userId"`
	DifferentLocation bool `json:"differentLocation"`
	IsNewDevice       bool `json:"isNewDevice"`
}

// SendVerifyCodeRequest 验证码请求
type SendVerifyCodeRequest struct {
	Identifier string `json:"identifier" binding:"required,min=3,max=50"`
	VerifyType string `json:"verifyType" binding:"required,oneof=EMAIL PHONE"`
	Action     string `json:"action" binding:"required,oneof=LOGIN REGISTER"`
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

// UserResponse 用户响应
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Phone    string `json:"phone"`
	Status   int    `json:"status"`
}

// ToResponse 转换为响应对象
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Phone:    u.Phone,
		Status:   u.Status,
	}
}
