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

// InitDatabase 初始化数据库连接
func InitDatabase(config *config.DatabaseConfig) *gorm.DB {
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

	return db
}

// InitRedis 初始化Redis连接
func InitRedis(config *config.RedisConfig) (*redis.Client, error) {
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

	return client, nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	config := config.GetConfig()
	return InitDatabase(&config.Database)
}

// GetRedis 获取Redis连接
func GetRedis() *redis.Client {
	config := config.GetConfig()
	client, err := InitRedis(&config.Redis)
	if err != nil {
		return nil
	}
	return client
}
