package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/LandcLi/LandcLogFace"
	"github.com/LandcLi/LandcUserManager/utility/client"
	"gopkg.in/yaml.v2"
)

// ThirdPartyGlobalConfig 库对外暴露的通用配置
// 用户只需在ThirdParty节点下按类型配置参数，无需感知内部Config
type ThirdPartyGlobalConfig struct {
	ThirdParty map[string]map[string]interface{} `mapstructure:"third_party"` // key=适配器类型（如"wechat"），value=该类型的配置参数
}

// Config 配置结构
type Config struct {
	Server       ServerConfig           `yaml:"server"`
	Database     DatabaseConfig         `yaml:"database"`
	Redis        RedisConfig            `yaml:"redis"`
	JWT          JWTConfig              `yaml:"jwt"`
	Email        client.EmailConfig     `yaml:"email"`
	SMS          SMSConfig              `yaml:"sms"`
	Log          LandcLogFace.LogConfig `yaml:"logger"`
	Cache        CacheConfig            `yaml:"cache"`
	Password     PasswordConfig         `yaml:"password"`
	VerifyCode   VerifyCodeConfig       `yaml:"verify_code"`
	HTTPClient   HTTPClientConfig       `yaml:"http_client"`
	LoginHistory LoginHistoryConfig     `yaml:"login_history"`
	RateLimit    RateLimitConfig        `yaml:"rate_limit"`
	ThridParty   ThirdPartyGlobalConfig `yaml:",inline"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port             int    `yaml:"port"`
	Host             string `yaml:"host"`
	ReadTimeout      int    `yaml:"read_timeout"`
	WriteTimeout     int    `yaml:"write_timeout"`
	GinMode          string `yaml:"gin_mode"`
	EnableCORS       bool   `yaml:"enable_cors"`
	UserDefaultRoute bool   `yaml:"user_default_route"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DSN             string `yaml:"dsn"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string `yaml:"addr"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`      // 连接池大小
	MinIdleConns int    `yaml:"min_idle_conns"` // 最小空闲连接数
	MaxRetries   int    `yaml:"max_retries"`    // 最大重试次数
	DialTimeout  int    `yaml:"dial_timeout"`   // 连接超时（秒）
	ReadTimeout  int    `yaml:"read_timeout"`   // 读取超时（秒）
	WriteTimeout int    `yaml:"write_timeout"`  // 写入超时（秒）
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey      string `yaml:"secret_key"`
	PrivateKeyPath string `yaml:"private_key_path"`
	PublicKeyPath  string `yaml:"public_key_path"`
	ExpiresIn      int    `yaml:"expires_in"`
	RefreshIn      int    `yaml:"refresh_in"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	ApiKey    string `yaml:"api_key"`
	ApiSecret string `yaml:"api_secret"`
	SignName  string `yaml:"sign_name"`
	Template  string `yaml:"template"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	UserCacheExpire int `yaml:"user_cache_expire"` // 用户缓存过期时间（秒）
}

// PasswordConfig 密码配置
type PasswordConfig struct {
	BcryptCost int `yaml:"bcrypt_cost"` // bcrypt加密成本
}

// VerifyCodeConfig 验证码配置
type VerifyCodeConfig struct {
	Length     int `yaml:"length"`      // 验证码长度
	ExpireTime int `yaml:"expire_time"` // 验证码过期时间（秒）
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	Timeout int `yaml:"timeout"` // 超时时间（秒）
}

// LoginHistoryConfig 登录历史配置
type LoginHistoryConfig struct {
	RetentionDays int `yaml:"retention_days"` // 登录历史保留天数
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	VerifyCodeInterval  int `yaml:"verify_code_interval"`    // 验证码发送间隔（秒）
	MaxVerifyCodePerDay int `yaml:"max_verify_code_per_day"` // 每天最多发送验证码次数
}

// 全局配置实例
var globalConfig *Config

// SetGlobalConfig 设置全局配置
func SetGlobalConfig(config *Config) {
	globalConfig = config
}

// LoadConfig 加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 读取配置文件
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置
	config := &Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置全局配置
	globalConfig = config

	return config, nil
}

// GetConfig 获取配置
func GetConfig() *Config {
	if globalConfig == nil {
		// 尝试从默认路径加载配置
		configPath := "config.yml"
		if _, err := os.Stat(configPath); err == nil {
			config, err := LoadConfig(configPath)
			if err == nil {
				return config
			}
		}

		// 加载失败，返回默认配置
		panic("配置加载失败")
	}

	return globalConfig
}
