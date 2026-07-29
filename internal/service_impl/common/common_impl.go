package common

import (
	"context"
	"fmt"
	"time"

	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/tools/generate"
	"github.com/LandcLi/landc-go/tools/mail"

	"github.com/LandcLi/LandcUserManager/model"
	lumerrors "github.com/LandcLi/LandcUserManager/pkg/errors"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/LandcUserManager/utility/client"
	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/LandcLi/LandcUserManager/utility/logger"

	constants "github.com/LandcLi/LandcUserManager/pkg/consts"
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
	case constants.VerifyTypeEmail:
		key = fmt.Sprintf(constants.CacheKeyVerifyCode, "email", req.Identifier, req.Action)
	case constants.VerifyTypePhone:
		key = fmt.Sprintf(constants.CacheKeyVerifyCode, "phone", req.Identifier, req.Action)
	default:
		return lumerrors.ErrInvalidVerifyType
	}

	c := framecache.GetCache()
	if c == nil {
		logger.Errorf(ctx, "缓存未初始化")
		return lumerrors.ErrVerifyCodeSendFailed
	}

	if err := c.Set(ctx, key, verifyCode, 5*time.Minute); err != nil {
		logger.Errorf(ctx, "存储验证码到缓存失败: %v", err)
		return lumerrors.ErrVerifyCodeSendFailed
	}

	switch req.VerifyType {
	case constants.VerifyTypeEmail:
		mailClient := client.GetMailClient()
		if mailClient == nil {
			return lumerrors.ErrMailClientNotInit
		}
		err := s.sendMail(req.Identifier, "验证码", fmt.Sprintf("您的验证码是: %s，5分钟内有效", verifyCode))
		if err != nil {
			logger.Errorf(ctx, "发送邮件验证码失败: %v", err)
			return lumerrors.ErrVerifyCodeSendFailed
		}
	case constants.VerifyTypePhone:
		if err := s.sendSMS(req.Identifier, verifyCode); err != nil {
			logger.Errorf(ctx, "发送短信验证码失败: %v", err)
			return lumerrors.ErrVerifyCodeSendFailed
		}
	}

	return nil
}

// CheckVerifyCode 验证验证码
func (s *commonServiceImpl) CheckVerifyCode(ctx context.Context, identifier string, verifyType string, verifyCode string, action string) error {
	var key string
	switch verifyType {
	case constants.VerifyTypeEmail:
		key = fmt.Sprintf(constants.CacheKeyVerifyCode, "email", identifier, action)
	case constants.VerifyTypePhone:
		key = fmt.Sprintf(constants.CacheKeyVerifyCode, "phone", identifier, action)
	default:
		return lumerrors.ErrInvalidVerifyType
	}

	c := framecache.GetCache()
	if c == nil {
		logger.Errorf(ctx, "缓存未初始化")
		return lumerrors.ErrVerifyCodeVerifyFailed
	}

	storedCode, err := c.Get(ctx, key)
	if err != nil {
		logger.Errorf(ctx, "从缓存获取验证码失败: %v", err)
		return lumerrors.ErrVerifyCodeExpired
	}

	if storedCode != verifyCode {
		return lumerrors.ErrVerifyCodeWrong
	}

	if err := c.Delete(ctx, key); err != nil {
		logger.Warnf(ctx, "删除已使用验证码失败: %v", err)
	}

	return nil
}

func (s *commonServiceImpl) sendMail(to string, subject string, body string) error {
	mailClient := client.GetMailClient()
	if mailClient == nil {
		return lumerrors.ErrMailClientNotInit
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
		return lumerrors.ErrSMSConfigNotSet
	}
	// TODO: implement SMS sending logic (planned feature)
	logger.Infof(context.Background(), "SMS sending not yet implemented: phone=%s, code=%s", phone, code)
	return nil
}
