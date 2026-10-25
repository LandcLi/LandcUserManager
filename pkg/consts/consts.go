// Package constants 定义项目全局常量，消除魔法字符串
package constants

// ---- 认证类型 ----
const (
	AuthTypePassword   = "PASSWORD"
	AuthTypeVerifyCode = "VERIFY_CODE"
	AuthTypeThirdParty = "THIRD_PARTY"
)

// ---- 验证类型 ----
const (
	VerifyTypeEmail    = "EMAIL"
	VerifyTypePhone    = "PHONE"
	VerifyTypeUsername = "USERNAME"
)

// ---- 验证码操作类型 ----
const (
	ActionLogin    = "login"
	ActionRegister = "register"
	ActionReset    = "reset"
)

// ---- 缓存 Key 格式 ----
const (
	CacheKeyUserByID       = "user:id:%d"           // user:id:%d
	CacheKeyUserByUsername = "user:username:%s"     // user:username:%s
	CacheKeyUserByEmail    = "user:email:%s"        // user:email:%s
	CacheKeyUserByPhone    = "user:phone:%s"        // user:phone:%s
	CacheKeyVerifyCode     = "verify_code:%s:%s:%s" // verify_code:{type}:{identifier}:{action}
)
