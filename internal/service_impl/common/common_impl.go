package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/tools/generate"
	"github.com/LandcLi/landc-go/tools/mail"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/LandcUserManager/utility/client"
	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/LandcLi/LandcUserManager/utility/logger"
)

type commonServiceImpl struct{}

func NewCommonService() *commonServiceImpl {
	return &commonServiceImpl{}
}

func init() {
	di.Provide[service.CommonService]("common.service", NewCommonService())
}

// SendVerifyCode 发送验证码
func (s *commonServiceImpl) SendVerifyCode(ctx context.Context, req *model.SendVerifyCodeRequest) error {
	verifyCode := generate.MustCaptcha(6)

	var key string
	switch req.VerifyType {
	case "EMAIL":
		key = fmt.Sprintf("verify_code:email:%s:%s", req.Identifier, req.Action)
	case "PHONE":
		key = fmt.Sprintf("verify_code:phone:%s:%s", req.Identifier, req.Action)
	default:
		return errors.New("无效的验证类型")
	}

	c := framecache.GetCache()
	if c == nil {
		logger.Errorf(ctx, "缓存未初始化")
		return errors.New("验证码发送失败")
	}

	if err := c.Set(ctx, key, verifyCode, 5*time.Minute); err != nil {
		logger.Errorf(ctx, "存储验证码到缓存失败: %v", err)
		return errors.New("验证码发送失败")
	}

	switch req.VerifyType {
	case "EMAIL":
		mailClient := client.GetMailClient()
		if mailClient == nil {
			return errors.New("邮件客户端未初始化")
		}
		err := s.sendMail(req.Identifier, "验证码", fmt.Sprintf("您的验证码是: %s，5分钟内有效", verifyCode))
		if err != nil {
			logger.Errorf(ctx, "发送邮件验证码失败: %v", err)
			return errors.New("验证码发送失败")
		}
	case "PHONE":
		if err := s.sendSMS(req.Identifier, verifyCode); err != nil {
			logger.Errorf(ctx, "发送短信验证码失败: %v", err)
			return errors.New("验证码发送失败")
		}
	}

	return nil
}

// CheckVerifyCode 验证验证码
func (s *commonServiceImpl) CheckVerifyCode(ctx context.Context, identifier string, verifyType string, verifyCode string, action string) error {
	var key string
	switch verifyType {
	case "EMAIL":
		key = fmt.Sprintf("verify_code:email:%s:%s", identifier, action)
	case "PHONE":
		key = fmt.Sprintf("verify_code:phone:%s:%s", identifier, action)
	default:
		return errors.New("无效的验证类型")
	}

	c := framecache.GetCache()
	if c == nil {
		logger.Errorf(ctx, "缓存未初始化")
		return errors.New("验证码验证失败")
	}

	storedCode, err := c.Get(ctx, key)
	if err != nil {
		logger.Errorf(ctx, "从缓存获取验证码失败: %v", err)
		return errors.New("验证码已过期或不存在")
	}

	if storedCode != verifyCode {
		return errors.New("验证码错误")
	}

	_ = c.Delete(ctx, key)

	return nil
}

func (s *commonServiceImpl) sendMail(to string, subject string, body string) error {
	mailClient := client.GetMailClient()
	if mailClient == nil {
		return errors.New("邮件客户端未初始化")
	}

	return mailClient.Send(&mail.Email{
		From:    config.GetConfig().Email.From,
		To:      []string{to},
		Subject: subject,
		Body:    body,
	})
}

func (s *commonServiceImpl) sendSMS(phone string, code string) error {
	if config.GetConfig() == nil || config.GetConfig().SMS.ApiKey == "" {
		return errors.New("短信配置未设置")
	}
	// todo 实现短信发送逻辑
	return nil
}
