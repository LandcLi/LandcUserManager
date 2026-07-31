package config

import (
	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	"github.com/LandcLi/landc-go/log/facade"

	"github.com/LandcLi/LandcUserManager/utility/client"
)

// ---- 类型定义（仅供业务代码类型化访问） ----
// 实际配置加载由框架全权负责，GetConfig() 从框架 Config 中提取并转回类型化结构。
// 字段 json tag 必须与 config.yaml 中的 key 对应（UnmarshalOther 基于 JSON 中转）。

type ThirdPartyGlobalConfig struct {
	ThirdParty map[string]map[string]interface{} `yaml:"third_party" json:"third_party"`
}

type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	JWT          JWTConfig
	Email        client.EmailConfig
	SMS          SMSConfig
	Log          facade.LogConfig
	Cache        CacheConfig
	Password     PasswordConfig
	VerifyCode   VerifyCodeConfig
	HTTPClient   HTTPClientConfig
	LoginHistory LoginHistoryConfig
	RateLimit    RateLimitConfig
	ThirdParty   ThirdPartyGlobalConfig
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

// GetConfig 从框架全局配置中提取 LUM 类型化配置。
// 配置由框架全权加载和管理，此函数仅做类型转换。
func GetConfig() *Config {
	fc := frameconfig.GetConfig()
	if fc == nil {
		panic("框架配置未初始化，请先调用 frameconfig.InitGlobalConfigWithPath")
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
	_ = fc.UnmarshalOther("third_party", &cfg.ThirdParty)

	return cfg
}
