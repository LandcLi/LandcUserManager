package db

import (
	"context"
	"fmt"
	"time"

	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var globalDB *gorm.DB
var globalRedis *redis.Client

// InitDatabase 初始化数据库连接
func InitDatabase(config *config.DatabaseConfig) *gorm.DB {
	if globalDB != nil {
		return globalDB
	}

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(config.DSN), gormConfig)
	if err != nil {
		panic(fmt.Sprintf("连接数据库失败: %v", err))
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("获取数据库连接失败: %v", err))
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	// 赋值给全局变量
	globalDB = db

	return db
}

// InitRedis 初始化Redis连接 (optional: if Addr is empty, Redis will not be initialized)
func InitRedis(config *config.RedisConfig) (*redis.Client, error) {
	if globalRedis != nil {
		return globalRedis, nil
	}

	// Skip if Redis is not configured
	if config == nil || config.Addr == "" {
		return nil, nil
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("连接Redis失败: %v", err)
	}

	// 赋值给全局变量
	globalRedis = client

	return client, nil
}

// InitDatabaseWithDB 直接传入db客户端
func InitDatabaseWithDB(db *gorm.DB) {
	globalDB = db
}

// InitRedisWithClient 直接传入redis客户端
func InitRedisWithClient(redisClient *redis.Client) {
	globalRedis = redisClient
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	if globalDB == nil {
		config := config.GetConfig()
		globalDB = InitDatabase(&config.Database)
	}
	return globalDB
}

// GetRedis 获取Redis连接 (may return nil if Redis is not configured)
func GetRedis() *redis.Client {
	if globalRedis == nil {
		cfg := config.GetConfig()
		if cfg == nil || cfg.Redis.Addr == "" {
			return nil
		}
		client, err := InitRedis(&cfg.Redis)
		if err != nil {
			return nil
		}
		globalRedis = client
	}
	return globalRedis
}
