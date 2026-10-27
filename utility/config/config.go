package config

import (
	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	"github.com/LandcLi/landc-go/log/facade"

	"github.com/LandcLi/LandcUserManager/utility/client"
)

// ---- 类型定义（仅供业务代码类型化访问） ----
// 实际配置加载由框架全权负责，GetConfig() 从框架 Config 中提取并转回类型化结构。
// 字段 json tag 必须与 config.yaml 中的 key 对应（UnmarshalOther 基于 JSON 中转）。

type ThirdPartyGlobalConfig = map[string]map[string]interface{}

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	JWT           JWTConfig
	Email         client.EmailConfig
	SMS           SMSConfig
	Log           facade.LogConfig
	Cache         CacheConfig
	Password      PasswordConfig
	VerifyCode    VerifyCodeConfig
	HTTPClient    HTTPClientConfig
	LoginHistory  LoginHistoryConfig
	RateLimit     RateLimitConfig
	LoginSecurity LoginSecurityConfig
	ServiceAuth   ServiceAuthConfig
	ThirdParty    ThirdPartyGlobalConfig
}

// ServiceAuthConfig 服务间调用的内部凭证。
//
// 背景：LandcSSO 需要从本服务读取用户档案（OIDC /userinfo 用令牌里的 sub 查档案），
// 那是没有「当前登录用户 token」的服务间调用，走用户 JWT 无处可带。给一个共享密钥，
// 服务间调用带 X-Service-Key 即可通过鉴权；外部客户端没有这个密钥，仍必须带 JWT。
//
// Key 为空表示不启用：此时所有非公开路径一律要求 JWT（不会因为漏配而放开）。
type ServiceAuthConfig struct {
	Key string `yaml:"key" json:"key"`
}

type ServerConfig struct {
	Port             int
	Host             string
	ReadTimeout      int
	WriteTimeout     int
	GinMode          string
	EnableCORS       bool
	UserDefaultRoute bool
	SecurityHeaders  SecurityHeadersConfig
}

// SecurityHeadersConfig 安全响应头开关（取自 other.server.security_headers）。
//
// 历史状态：config.example.yml 里一直写着 server.security_headers.enabled: true，
// 但代码里既没有对应字段也没有挂中间件 —— 配置在说谎，响应里一个安全头都没有。
// 框架本身提供了 api/middleware/gin.SecurityHeaders，只是从没被调用。
//
// HSTS 默认关闭（0）：内网常以 HTTP 部署，发 Strict-Transport-Security 会让浏览器
// 把该域名强制升级到 HTTPS，反而把服务访问搞坏。确认走 HTTPS 后再显式开启。
type SecurityHeadersConfig struct {
	Enabled            bool   `yaml:"enabled" json:"enabled"`
	HSTSMaxAge         int    `yaml:"hsts_max_age" json:"hsts_max_age"`
	FrameDeny          *bool  `yaml:"frame_deny" json:"frame_deny"`
	ContentTypeNoSniff *bool  `yaml:"content_type_nosniff" json:"content_type_nosniff"`
	ReferrerPolicy     string `yaml:"referrer_policy" json:"referrer_policy"`
}
type DatabaseConfig struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
}
type JWTConfig struct {
	Secret     string `yaml:"secret" json:"secret"`
	ExpireTime string `yaml:"expire_time" json:"expire_time"`
	Issuer     string `yaml:"issuer" json:"issuer"`
}
type SMSConfig struct {
	ApiKey    string `yaml:"api_key" json:"api_key"`
	ApiSecret string `yaml:"api_secret" json:"api_secret"`
	SignName  string `yaml:"sign_name" json:"sign_name"`
	Template  string `yaml:"template" json:"template"`
}
type CacheConfig struct {
	UserCacheExpire int `yaml:"user_cache_expire" json:"user_cache_expire"`
}
type PasswordConfig struct {
	BcryptCost int `yaml:"bcrypt_cost" json:"bcrypt_cost"`
}
type VerifyCodeConfig struct {
	Length     int `yaml:"length" json:"length"`
	ExpireTime int `yaml:"expire_time" json:"expire_time"`
}
type HTTPClientConfig struct {
	Timeout int `yaml:"timeout" json:"timeout"`
}
type LoginHistoryConfig struct {
	RetentionDays int `yaml:"retention_days" json:"retention_days"`
}
type RateLimitConfig struct {
	VerifyCodeInterval  int `yaml:"verify_code_interval" json:"verify_code_interval"`
	MaxVerifyCodePerDay int `yaml:"max_verify_code_per_day" json:"max_verify_code_per_day"`
}

// LoginSecurityConfig 登录安全策略（失败计数 + 账号锁定）。
//
// 三个数值都可配：max_failed_attempts 决定「多少次连续失败算异常」，
// failure_window_minutes 决定「多久没失败就清零」，lock_minutes 决定「锁多久」。
// max_failed_attempts <= 0 表示关闭锁定（仅计数不拦截）。
type LoginSecurityConfig struct {
	Enabled              bool `yaml:"enabled" json:"enabled"`
	MaxFailedAttempts    int  `yaml:"max_failed_attempts" json:"max_failed_attempts"`
	LockMinutes          int  `yaml:"lock_minutes" json:"lock_minutes"`
	FailureWindowMinutes int  `yaml:"failure_window_minutes" json:"failure_window_minutes"`
}

// 默认登录安全策略：连续失败 5 次锁 15 分钟，失败计数窗口 15 分钟。
const (
	defaultMaxFailedAttempts    = 5
	defaultLockMinutes          = 15
	defaultFailureWindowMinutes = 15
)

// defaultLoginSecurity 返回默认策略。
// 注意：必须「先填默认值再反序列化」—— YAML 里省略的键不会覆盖已有值，
// 否则 enabled 会被静默置 false（v0.8.5 的 use_default_routes 就踩过这个坑）。
func defaultLoginSecurity() LoginSecurityConfig {
	return LoginSecurityConfig{
		Enabled:              true,
		MaxFailedAttempts:    defaultMaxFailedAttempts,
		LockMinutes:          defaultLockMinutes,
		FailureWindowMinutes: defaultFailureWindowMinutes,
	}
}

// defaultLoginHistoryRetentionDays 登录历史默认保留天数。
// 与 config.example.yml 保持一致；<=0 表示不做清理（不推荐）。
const defaultLoginHistoryRetentionDays = 90

func defaultLoginHistory() LoginHistoryConfig {
	return LoginHistoryConfig{RetentionDays: defaultLoginHistoryRetentionDays}
}

// 验证码默认值。与 tools/verifycode 的内建默认值一致 ——
// 之所以在这一层也给一遍，是因为下面这些值现在会显式传给 Manager，
// 传 0 会真的生成「0 位验证码」而不是回落默认。
const (
	defaultVerifyCodeLength   = 6
	defaultVerifyCodeExpire   = 300 // 秒
	defaultVerifyCodeInterval = 60  // 秒
	defaultVerifyCodeDailyMax = 10  // 次/天
)

func defaultVerifyCode() VerifyCodeConfig {
	return VerifyCodeConfig{Length: defaultVerifyCodeLength, ExpireTime: defaultVerifyCodeExpire}
}

func defaultRateLimit() RateLimitConfig {
	return RateLimitConfig{
		VerifyCodeInterval:  defaultVerifyCodeInterval,
		MaxVerifyCodePerDay: defaultVerifyCodeDailyMax,
	}
}

// GetConfig 从框架全局配置中提取 LUM 类型化配置。
// 配置由框架全权加载和管理，此函数仅做类型转换。
// 框架配置未初始化时 panic —— 适合 main 启动路径上的「必须有配置」调用点。
func GetConfig() *Config {
	cfg := TryGetConfig()
	if cfg == nil {
		panic("框架配置未初始化，请先调用 frameconfig.InitGlobalConfigWithPath")
	}
	return cfg
}

// TryGetConfig 与 GetConfig 语义相同，但配置未初始化时返回 nil 而不 panic。
//
// 存在的意义：有两条路径不能假设配置已就绪 ——
//   - 单元测试用 mock 装配 service，不 bootstrap 框架配置；
//   - 本系统被当库嵌入到宿主进程时，宿主可能只装了 DB 没装配置。
//
// 这些场景下「拿不到配置」应该是降级（用默认策略），而不是整条链路 panic。
func TryGetConfig() *Config {
	fc := frameconfig.GetConfig()
	if fc == nil {
		return nil
	}
	return extractLUMConfig(fc)
}

func extractLUMConfig(fc *frameconfig.Config) *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host: fc.Server.Addr, Port: fc.Server.Port,
		},
		Database: DatabaseConfig{
			DSN: fc.Database.DSN, MaxOpenConns: fc.Database.MaxOpenConns,
			MaxIdleConns: fc.Database.MaxIdleConns, ConnMaxLifetime: fc.Database.ConnMaxLifetime,
		},
		Redis: RedisConfig{
			Addr: fc.Redis.Addr, Password: fc.Redis.Password,
			DB: fc.Redis.DB, PoolSize: fc.Redis.PoolSize,
		},
		// 先填默认值，再由 UnmarshalOther 覆盖配置里显式给出的键。
		LoginSecurity: defaultLoginSecurity(),
		LoginHistory:  defaultLoginHistory(),
		VerifyCode:    defaultVerifyCode(),
		RateLimit:     defaultRateLimit(),
	}

	// other.server：本系统在框架 server 段之外的扩展项。
	// gin_mode 此前只声明了字段、从未赋值，等于配了不生效。
	var serverOther struct {
		GinMode         string               `json:"gin_mode"`
		SecurityHeaders SecurityHeadersConfig `json:"security_headers"`
	}
	_ = fc.UnmarshalOther("server", &serverOther)
	cfg.Server.GinMode = serverOther.GinMode
	cfg.Server.SecurityHeaders = serverOther.SecurityHeaders

	// 自定义配置通过 UnmarshalOther 类型化（基于 JSON 中转）
	_ = fc.UnmarshalOther("jwt", &cfg.JWT)
	_ = fc.UnmarshalOther("email", &cfg.Email)
	_ = fc.UnmarshalOther("sms", &cfg.SMS)
	_ = fc.UnmarshalOther("log", &cfg.Log)
	_ = fc.UnmarshalOther("cache", &cfg.Cache)
	_ = fc.UnmarshalOther("password", &cfg.Password)
	_ = fc.UnmarshalOther("verify_code", &cfg.VerifyCode)
	_ = fc.UnmarshalOther("http_client", &cfg.HTTPClient)
	_ = fc.UnmarshalOther("login_history", &cfg.LoginHistory)
	_ = fc.UnmarshalOther("rate_limit", &cfg.RateLimit)
	_ = fc.UnmarshalOther("login_security", &cfg.LoginSecurity)
	_ = fc.UnmarshalOther("service_auth", &cfg.ServiceAuth)
	_ = fc.UnmarshalOther("third_party", &cfg.ThirdParty)

	// 历史遗留：password.bcrypt_cost 以前根本没被读取，代码里写死 DefaultCost，
	// 配了也不生效（测试想调低 cost 提速同样无效）。这里统一兜底。
	if cfg.Password.BcryptCost <= 0 {
		cfg.Password.BcryptCost = 10 // 与 bcrypt.DefaultCost 一致
	}

	return cfg
}
