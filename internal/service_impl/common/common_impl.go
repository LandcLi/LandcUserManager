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

// SendVerifyCode 发送验证码（含发送间隔与每日次数限流）
func (s *commonServiceImpl) SendVerifyCode(ctx context.Context, req *model.SendVerifyCodeRequest) error {
	var key string
	switch req.VerifyType {
	case constants.VerifyTypeEmail:
		key = fmt.Sprintf(constants.CacheKeyVerifyCode, "email", req.Identifier, req.Action)
	case constants.VerifyTypePhone:
		key = fmt.Sprintf(constants.CacheKeyVerifyCode, "phone", req.Identifier, req.Action)
	default:
		return lumerrors.ErrInvalidVerifyType
	}

	c := framecache.GetCacheFrom(ctx)
	if c == nil {
		logger.Errorf(ctx, "缓存未初始化")
		return lumerrors.ErrVerifyCodeSendFailed
	}

	rate := config.GetConfig().RateLimit

	// 限流：发送间隔内不允许重复发送
	intervalKey := key + ":interval"
	exists, err := c.Exists(ctx, intervalKey)
	if err != nil {
		logger.Warnf(ctx, "check verify code interval failed: %v", err)
	} else if exists {
		return lumerrors.ErrTooManyRequests
	}

	// 限流：每日发送次数上限
	dayKey := key + ":daily"
	dayCount, err := s.getDailyCount(ctx, c, dayKey)
	if err != nil {
		logger.Warnf(ctx, "check daily count failed: %v", err)
	} else if rate.MaxVerifyCodePerDay > 0 && dayCount >= rate.MaxVerifyCodePerDay {
		return lumerrors.ErrDailyLimitExceeded
	}

	verifyCode := generate.MustCaptcha(6)

	if err := c.Set(ctx, key, verifyCode, 5*time.Minute); err != nil {
		logger.Errorf(ctx, "存储验证码到缓存失败: %v", err)
		return lumerrors.ErrVerifyCodeSendFailed
	}

	// 设置发送间隔标记
	interval := time.Duration(rate.VerifyCodeInterval) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}
	if err := c.Set(ctx, intervalKey, "1", interval); err != nil {
		logger.Warnf(ctx, "set verify code interval marker failed: %v", err)
	}

	// 更新每日计数（TTL 24 小时）
	if err := s.incrementDailyCount(ctx, c, dayKey, dayCount); err != nil {
		logger.Warnf(ctx, "update daily count failed: %v", err)
	}

	switch req.VerifyType {
	case constants.VerifyTypeEmail:
		mailClient := client.GetMailClient()
		if mailClient == nil {
			return lumerrors.ErrMailClientNotInit
		}
		err := s.sendMail(ctx, req.Identifier, "验证码", fmt.Sprintf("您的验证码是: %s，5分钟内有效", verifyCode))
		if err != nil {
			logger.Errorf(ctx, "发送邮件验证码失败: %v", err)
			return lumerrors.ErrVerifyCodeSendFailed
		}
	case constants.VerifyTypePhone:
		if err := s.sendSMS(ctx, req.Identifier, verifyCode); err != nil {
			logger.Errorf(ctx, "发送短信验证码失败: %v", err)
			return lumerrors.ErrVerifyCodeSendFailed
		}
	}

	return nil
}

// getDailyCount 读取当日已发送次数
func (s *commonServiceImpl) getDailyCount(ctx context.Context, c interface {
	Get(context.Context, string) (string, error)
}, dayKey string) (int, error) {
	v, err := c.Get(ctx, dayKey)
	if err != nil || v == "" {
		return 0, nil
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return 0, err
	}
	return n, nil
}

// incrementDailyCount 递增每日计数（基于 Get+Set，非原子；高并发下可能有误差，但不影响安全上限兜底）
func (s *commonServiceImpl) incrementDailyCount(ctx context.Context, c interface {
	Set(context.Context, string, interface{}, time.Duration) error
}, dayKey string, current int) error {
	return c.Set(ctx, dayKey, fmt.Sprintf("%d", current+1), 24*time.Hour)
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

	c := framecache.GetCacheFrom(ctx)
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

func (s *commonServiceImpl) sendMail(ctx context.Context, to string, subject string, body string) error {
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

func (s *commonServiceImpl) sendSMS(ctx context.Context, phone string, code string) error {
	if config.GetConfig() == nil || config.GetConfig().SMS.ApiKey == "" {
		return lumerrors.ErrSMSConfigNotSet
	}
	// TODO: implement SMS sending logic (planned feature)
	logger.Infof(ctx, "SMS sending not yet implemented: phone=%s, code=%s", phone, code)
	return nil
}
