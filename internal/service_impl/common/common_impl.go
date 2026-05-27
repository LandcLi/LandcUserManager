package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/LandcUserManager/utility/client"
	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/LandcLi/LandcUserManager/utility/db"
	"github.com/LandcLi/LandcUserManager/utility/logger"
	"github.com/LandcLi/landc-go/tools/generate"
	"github.com/LandcLi/landc-go/tools/mail"
	"github.com/go-redis/redis/v8"
)

// commonServiceImpl 通用服务实现
type commonServiceImpl struct {
	config *config.Config
}

// NewCommonService 创建通用服务实例
func NewCommonService() *commonServiceImpl {
	return &commonServiceImpl{
		config: config.GetConfig(),
	}
}

func init() {
	if service.GetCommonService() == nil {
		service.RegisterCommonService(NewCommonService())
	}
}

// SendVerifyCode 发送验证码
func (s *commonServiceImpl) SendVerifyCode(ctx context.Context, req *model.SendVerifyCodeRequest) error {
	// 生成6位数字验证码
	verifyCode := generate.MustCaptcha(6)

	// 构建Redis键
	var key string
	switch req.VerifyType {
	case "EMAIL":
		key = fmt.Sprintf("verify_code:email:%s:%s", req.Identifier, req.Action)
	case "PHONE":
		key = fmt.Sprintf("verify_code:phone:%s:%s", req.Identifier, req.Action)
	default:
		return errors.New("无效的验证类型")
	}

	// 存储验证码到Redis，设置5分钟过期
	redisClient := db.GetRedis()
	if redisClient == nil {
		logger.Errorf(ctx, "Redis客户端未初始化")
		return errors.New("验证码发送失败")
	}

	err := redisClient.Set(ctx, key, verifyCode, 5*time.Minute).Err()
	if err != nil {
		logger.Errorf(ctx, "存储验证码到Redis失败: %v", err)
		return errors.New("验证码发送失败")
	}

	// 发送验证码
	switch req.VerifyType {
	case "EMAIL":
		mailClient := client.GetMailClient()
		if mailClient == nil {
			return errors.New("邮件客户端未初始化")
		}
		err = s.sendMail(req.Identifier, "验证码", fmt.Sprintf("您的验证码是: %s，5分钟内有效", verifyCode))
	case "PHONE":
		err = s.sendSMS(req.Identifier, verifyCode)
	default:
		return errors.New("无效的验证类型")
	}

	if err != nil {
		logger.Errorf(ctx, "发送验证码失败: %v", err)
		return errors.New("验证码发送失败")
	}

	return nil
}

// CheckVerifyCode 验证验证码
func (s *commonServiceImpl) CheckVerifyCode(ctx context.Context, identifier string, verifyType string, verifyCode string, action string) error {
	// 构建Redis键
	var key string
	switch verifyType {
	case "EMAIL":
		key = fmt.Sprintf("verify_code:email:%s:%s", identifier, action)
	case "PHONE":
		key = fmt.Sprintf("verify_code:phone:%s:%s", identifier, action)
	default:
		return errors.New("无效的验证类型")
	}

	// 从Redis获取验证码
	redisClient := db.GetRedis()
	if redisClient == nil {
		logger.Errorf(ctx, "Redis客户端未初始化")
		return errors.New("验证码验证失败")
	}

	storedCode, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New("验证码已过期或不存在")
		}
		logger.Errorf(ctx, "从Redis获取验证码失败: %v", err)
		return errors.New("验证码验证失败")
	}

	// 验证验证码
	if storedCode != verifyCode {
		return errors.New("验证码错误")
	}

	// 验证码验证成功后删除
	_ = redisClient.Del(ctx, key).Err()

	return nil
}

// sendMail 发送邮件
func (s *commonServiceImpl) sendMail(to string, subject string, body string) error {
	mailClient := client.GetMailClient()
	if mailClient == nil {
		return errors.New("邮件客户端未初始化")
	}

	return mailClient.Send(&mail.Email{
		From:    s.config.Email.From,
		To:      []string{to},
		Subject: subject,
		Body:    body,
	})
}

// sendSMS 发送短信
func (s *commonServiceImpl) sendSMS(phone string, code string) error {
	if s.config == nil || s.config.SMS.ApiKey == "" {
		return errors.New("短信配置未设置")
	}
	// todo 实现短信发送逻辑
	return nil
}
