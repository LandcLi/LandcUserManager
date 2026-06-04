package config

import (
	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	"github.com/LandcLi/landc-go/log/facade"

	"github.com/LandcLi/LandcUserManager/utility/client"
)

// ---- 类型定义（仅供业务代码类型化访问） ----
// 实际配置加载由框架全权负责，GetConfig() 从框架 Config 中提取并转回类型化结构。

type ThirdPartyGlobalConfig struct {
	ThirdParty map[string]map[string]interface{} `yaml:"third_party"`
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
	ThridParty   ThirdPartyGlobalConfig
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
	SecretKey      string
	PrivateKeyPath string
	PublicKeyPath  string
	ExpiresIn      int
	RefreshIn      int
}
type SMSConfig struct {
	ApiKey    string
	ApiSecret string
	SignName  string
	Template  string
}
type CacheConfig struct {
	UserCacheExpire int
}
type PasswordConfig struct {
	BcryptCost int
}
type VerifyCodeConfig struct {
	Length     int
	ExpireTime int
}
type HTTPClientConfig struct {
	Timeout int
}
type LoginHistoryConfig struct {
	RetentionDays int
}
type RateLimitConfig struct {
	VerifyCodeInterval  int
	MaxVerifyCodePerDay int
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

	if fc.Other != nil {
		cfg.JWT, _ = fc.Other["jwt"].(JWTConfig)
		cfg.Email, _ = fc.Other["email"].(client.EmailConfig)
		cfg.SMS, _ = fc.Other["sms"].(SMSConfig)
		cfg.Log, _ = fc.Other["log"].(facade.LogConfig)
		cfg.Cache, _ = fc.Other["cache"].(CacheConfig)
		cfg.Password, _ = fc.Other["password"].(PasswordConfig)
		cfg.VerifyCode, _ = fc.Other["verify_code"].(VerifyCodeConfig)
		cfg.HTTPClient, _ = fc.Other["http_client"].(HTTPClientConfig)
		cfg.LoginHistory, _ = fc.Other["login_history"].(LoginHistoryConfig)
		cfg.RateLimit, _ = fc.Other["rate_limit"].(RateLimitConfig)
		cfg.ThridParty, _ = fc.Other["third_party"].(ThirdPartyGlobalConfig)
	}

	return cfg
}
