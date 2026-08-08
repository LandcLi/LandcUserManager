// Package errors 定义项目全局错误变量，支持 errors.Is() / errors.As() 判断
package lumerrors

import "errors"

// ---- 通用错误 ----
var (
	ErrInvalidAuthType   = errors.New("invalid auth type")
	ErrInvalidVerifyType = errors.New("invalid verify type")
)

// ---- 用户相关错误 ----
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserDisabled       = errors.New("user disabled")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrCreateUserFailed   = errors.New("create user failed")
	ErrUpdateUserFailed   = errors.New("update user info failed")
	ErrPasswordIncorrect  = errors.New("password incorrect")
	ErrOldPasswordWrong   = errors.New("old password wrong")
	ErrPasswordHashFailed = errors.New("password hash failed")
	ErrUpdatePasswordFail = errors.New("update password failed")
	ErrResetPasswordFail  = errors.New("reset password failed")
	ErrRegisterTokenFailed = errors.New("register token generation failed")
)

// ---- 限流相关错误 ----
var (
	ErrTooManyRequests       = errors.New("too many requests, please retry later")
	ErrDailyLimitExceeded    = errors.New("daily verify code limit exceeded")
)

// ---- 请求参数错误 ----
var (
	ErrInvalidRequestParams = errors.New("invalid request parameters")
)

// ---- 验证码相关错误 ----
var (
	ErrVerifyCodeSendFailed     = errors.New("verify code send failed")
	ErrVerifyCodeVerifyFailed   = errors.New("verify code verify failed")
	ErrVerifyCodeExpired        = errors.New("verify code expired or not exist")
	ErrVerifyCodeWrong          = errors.New("verify code wrong")
)

// ---- 邮件客户端错误 ----
var (
	ErrMailClientNotInit = errors.New("mail client not initialized")
	ErrSMSConfigNotSet   = errors.New("SMS config not set")
)

// ---- 第三方登录错误 ----
var (
	ErrThirdPartyNotSupported       = errors.New("unsupported third party login type")
	ErrThirdPartyTokenInvalid       = errors.New("third party token invalid or expired")
	ErrThirdPartyUserDisabled       = errors.New("third party user disabled")
	ErrThirdPartyFetchUserInfoFailed = errors.New("fetch third party user info failed")
)

// ---- 第三方API错误 ----
var (
	ErrBuildRequestFailed     = errors.New("build request failed")
	ErrCreateRequestFailed    = errors.New("create request failed")
	ErrFetchUserInfoFailed    = errors.New("fetch user info failed")
	ErrParseResponseFailed    = errors.New("parse response failed")
	ErrAPIError               = "API error: " // prefix, use with fmt.Errorf
)
