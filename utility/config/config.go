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
	ThirdParty    ThirdPartyGlobalConfig
}

type ServerConfig struct {
	Port             int
	Host             string
	ReadTimeout      int
	WriteTimeout     int
	GinMode          string
	EnableCORS       bool
	UserDefaultRoute bool
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
	}

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
	_ = fc.UnmarshalOther("third_party", &cfg.ThirdParty)

	// 历史遗留：password.bcrypt_cost 以前根本没被读取，代码里写死 DefaultCost，
	// 配了也不生效（测试想调低 cost 提速同样无效）。这里统一兜底。
	if cfg.Password.BcryptCost <= 0 {
		cfg.Password.BcryptCost = 10 // 与 bcrypt.DefaultCost 一致
	}

	return cfg
}
