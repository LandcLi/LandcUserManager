// Package errors 定义项目全局错误变量。
// 统一使用 landc-go api/core 的标准错误（带错误码），
// 使框架 ConvertFrameworkError 能正确映射业务错误码而非一律 500。
// 支持 errors.Is()/errors.As() 判断（返回同一包级变量指针，== 恒成立）。
package lumerrors

import (
	"fmt"
	"math"
	"time"

	"github.com/LandcLi/landc-go/api/core"
)

// ---- 通用错误 ----
var (
	ErrInvalidAuthType   = core.NewError(core.ErrorCodeInvalidParams, "invalid auth type")
	ErrInvalidVerifyType = core.NewError(core.ErrorCodeInvalidParams, "invalid verify type")
)

// ---- 请求参数错误 ----
var (
	ErrInvalidRequestParams = core.NewError(core.ErrorCodeInvalidParams, "invalid request parameters")
)

// ---- 用户相关错误 ----
var (
	ErrUserNotFound        = core.NewError(core.ErrorCodeNotFound, "user not found")
	ErrUserDisabled        = core.NewError(core.ErrorCodeForbidden, "user disabled")
	ErrUserAlreadyExists   = core.NewError(core.ErrorCodeConflict, "user already exists")
	ErrCreateUserFailed    = core.NewError(core.ErrorCodeDatabaseError, "create user failed")
	ErrUpdateUserFailed    = core.NewError(core.ErrorCodeDatabaseError, "update user info failed")
	ErrPasswordIncorrect   = core.NewError(core.ErrorCodeUnauthorized, "password incorrect")
	ErrOldPasswordWrong    = core.NewError(core.ErrorCodeUnauthorized, "old password wrong")
	ErrPasswordHashFailed  = core.NewError(core.ErrorCodeInternalServerError, "password hash failed")
	ErrUpdatePasswordFail  = core.NewError(core.ErrorCodeDatabaseError, "update password failed")
	ErrResetPasswordFail   = core.NewError(core.ErrorCodeDatabaseError, "reset password failed")
	ErrRegisterTokenFailed = core.NewError(core.ErrorCodeInternalServerError, "register token generation failed")
)

// ---- 登录安全错误 ----
var (
	// ErrInvalidCredentials 是登录场景对外统一的「凭据不正确」错误。
	//
	// 「用户不存在」与「密码错误」必须返回**同一个**错误值和同一段文案，
	// 否则攻击者可以用不同响应把用户名批量枚举出来（code 40400 vs 40100 一眼可辨）。
	// 需要区分时只看服务端日志，不要回给客户端。
	ErrInvalidCredentials = core.NewError(core.ErrorCodeUnauthorized, "invalid credentials")
)

// ErrorCodeAccountLocked 是「账号因连续登录失败被临时锁定」的业务错误码。
// 取自定义区间（60000-99999），客户端可据此弹出「请 X 分钟后再试」，
// 而不必去解析文案。
const ErrorCodeAccountLocked core.ErrorCode = core.ErrorCodeCustomMin + 429

// NewAccountLockedError 构造带剩余锁定时间的锁定错误。
// remaining <= 0 时按「已到期」处理，返回 nil 交回调用方（正常流程不该走到这一步）。
func NewAccountLockedError(remaining time.Duration) *core.Error {
	if remaining <= 0 {
		return nil
	}
	// 向上取整到分钟：剩 30 秒也应该提示「1 分钟」，不能显示「0 分钟」。
	minutes := int(math.Ceil(remaining.Minutes()))
	return &core.Error{
		Code:    ErrorCodeAccountLocked,
		Message: fmt.Sprintf("account locked due to too many failed login attempts, retry in %d minute(s)", minutes),
		Details: map[string]interface{}{"retry_after_seconds": int(math.Ceil(remaining.Seconds()))},
	}
}

// ---- 验证码相关错误 ----
var (
	ErrVerifyCodeSendFailed   = core.NewError(core.ErrorCodeExternalAPIError, "verify code send failed")
	ErrVerifyCodeVerifyFailed = core.NewError(core.ErrorCodeValidationFailed, "verify code verify failed")
	ErrVerifyCodeExpired      = core.NewError(core.ErrorCodeValidationFailed, "verify code expired or not exist")
	ErrVerifyCodeWrong        = core.NewError(core.ErrorCodeValidationFailed, "verify code wrong")
)

// ---- 邮件/短信客户端错误 ----
var (
	ErrMailClientNotInit = core.NewError(core.ErrorCodeInternalServerError, "mail client not initialized")
	ErrSMSConfigNotSet   = core.NewError(core.ErrorCodeInternalServerError, "SMS config not set")
)

// ---- 限流相关错误 ----
var (
	ErrTooManyRequests    = core.NewError(core.ErrorCodeServiceUnavailable, "too many requests, please retry later")
	ErrDailyLimitExceeded = core.NewError(core.ErrorCodeServiceUnavailable, "daily verify code limit exceeded")
)

// ---- 第三方登录错误 ----
var (
	ErrThirdPartyNotSupported        = core.NewError(core.ErrorCodeBadRequest, "unsupported third party login type")
	ErrThirdPartyTokenInvalid        = core.NewError(core.ErrorCodeUnauthorized, "third party token invalid or expired")
	ErrThirdPartyUserDisabled        = core.NewError(core.ErrorCodeForbidden, "third party user disabled")
	ErrThirdPartyFetchUserInfoFailed = core.NewError(core.ErrorCodeExternalAPIError, "fetch third party user info failed")
)

// ---- 第三方API错误 ----
var (
	ErrBuildRequestFailed  = core.NewError(core.ErrorCodeExternalAPIError, "build request failed")
	ErrCreateRequestFailed = core.NewError(core.ErrorCodeExternalAPIError, "create request failed")
	ErrFetchUserInfoFailed = core.NewError(core.ErrorCodeExternalAPIError, "fetch user info failed")
	ErrParseResponseFailed = core.NewError(core.ErrorCodeExternalAPIError, "parse response failed")
	// ErrAPIError 为错误前缀（配合 fmt.Errorf 使用），非独立错误值
	ErrAPIError = "API error: "
)
