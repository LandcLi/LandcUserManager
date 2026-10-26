package v1

import "github.com/LandcLi/landc-go/frame/pkg/meta"

type LoginRequest struct {
	meta.Meta `path:"/api/user/login" method:"POST"`
	// Identifier 标识（邮箱/手机/用户名）。PASSWORD/VERIFY_CODE 模式必填；THIRD_PARTY 模式不使用
	// （用户信息由 LUM 通过 AccessToken 调用 FetchUserInfo 获取），必填校验在 service 层按 AuthType 分支执行。
	Identifier string `json:"identifier" binding:"omitempty,min=3,max=50"`
	// AuthCode 密码/验证码。PASSWORD/VERIFY_CODE 模式必填；THIRD_PARTY 模式不使用
	// （openId 由 LUM 从 token 换取），个别 provider（如 wechat 二次校验）按约定可选传入。
	AuthCode string `json:"verifyCode" binding:"omitempty"` //密码/验证码（THIRD_PARTY 模式不使用）
	// AccessToken 第三方登录凭证（THIRD_PARTY 模式必填）。
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
	UserID            uint   `json:"userId"`
	DifferentLocation bool   `json:"differentLocation"`
	IsNewDevice       bool   `json:"isNewDevice"`
	Token             string `json:"token"`
	RefreshToken      string `json:"refreshToken"`
	// Username 用户名。
	//
	// 调用方（如 LandcSSO）签发自己的令牌时需要把用户名写进 claims。
	// 此前这里不返回用户名，SSO 只能给令牌填空串，导致下游拿到
	// userID 有效但 username 为空的令牌，按用户名做的逻辑全部失效。
	Username string `json:"username"`
}
