package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/tools/mail"
	toolsvc "github.com/LandcLi/landc-go/tools/verifycode"

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

// vcKey 构造验证码业务 key（frame/pkg/verifycode 内建 vc:code:/vc:interval:/vc:daily: 前缀，
// 业务侧只需保证 type/identifier/action 唯一）。
func (s *commonServiceImpl) vcKey(verifyType, identifier, action string) (string, error) {
	switch verifyType {
	case constants.VerifyTypeEmail:
		return fmt.Sprintf("email:%s:%s", identifier, action), nil
	case constants.VerifyTypePhone:
		return fmt.Sprintf("phone:%s:%s", identifier, action), nil
	default:
		return "", lumerrors.ErrInvalidVerifyType
	}
}

// vcManager 按配置构造验证码管理器。
//
// 为什么不再直接用 frame/pkg/verifycode.Generate：
// 那个封装把选项写死成内建默认值（6 位 / 5 分钟 / 60 秒间隔 / 10 次每天），
// 完全不读配置 —— 于是 verify_code.length、verify_code.expire_time、
// rate_limit.verify_code_interval、rate_limit.max_verify_code_per_day
// 这四项配了也不生效。选项恰好与默认值一致，所以改配置才会发现。
//
// 配置层已为这四项填好默认值（见 utility/config.defaultVerifyCode），
// 因此这里可以放心显式传参，不用担心传 0 生成「0 位验证码」。
func (s *commonServiceImpl) vcManager(ctx context.Context) (*toolsvc.Manager, error) {
	c := framecache.GetCacheFrom(ctx)
	if c == nil {
		return nil, lumerrors.ErrVerifyCodeSendFailed
	}

	cfg := config.TryGetConfig()
	if cfg == nil {
		// 配置未就绪时退回内建默认值
		return toolsvc.NewManager(framecache.AsToolsCache(c)), nil
	}

	opts := []toolsvc.Option{
		toolsvc.WithCodeLength(cfg.VerifyCode.Length),
		toolsvc.WithTTL(time.Duration(cfg.VerifyCode.ExpireTime) * time.Second),
		toolsvc.WithSendInterval(time.Duration(cfg.RateLimit.VerifyCodeInterval) * time.Second),
		toolsvc.WithDailyLimit(int64(cfg.RateLimit.MaxVerifyCodePerDay)),
	}
	return toolsvc.NewManager(framecache.AsToolsCache(c), opts...), nil
}

// SendVerifyCode 发送验证码。
// 生成、存储、发送间隔与每日上限由 tools/verifycode 承担（选项来自配置），
// 本方法只负责发送渠道。
func (s *commonServiceImpl) SendVerifyCode(ctx context.Context, req *model.SendVerifyCodeRequest) error {
	key, err := s.vcKey(req.VerifyType, req.Identifier, req.Action)
	if err != nil {
		return err
	}

	mgr, err := s.vcManager(ctx)
	if err != nil {
		logger.Errorf(ctx, "初始化验证码管理器失败: %v", err)
		return err
	}

	verifyCode, err := mgr.Generate(key)
	if err != nil {
		switch {
		case errors.Is(err, toolsvc.ErrIntervalLimited):
			return lumerrors.ErrTooManyRequests
		case errors.Is(err, toolsvc.ErrDailyLimited):
			return lumerrors.ErrDailyLimitExceeded
		default:
			logger.Errorf(ctx, "生成验证码失败: %v", err)
			return lumerrors.ErrVerifyCodeSendFailed
		}
	}

	expireMinutes := 5
	if cfg := config.TryGetConfig(); cfg != nil && cfg.VerifyCode.ExpireTime > 0 {
		expireMinutes = cfg.VerifyCode.ExpireTime / 60
	}

	switch req.VerifyType {
	case constants.VerifyTypeEmail:
		if err := s.sendMail(ctx, req.Identifier, "验证码",
			fmt.Sprintf("您的验证码是: %s，%d分钟内有效", verifyCode, expireMinutes)); err != nil {
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

// CheckVerifyCode 校验验证码（一次性：成功后删除，防重放）。
func (s *commonServiceImpl) CheckVerifyCode(ctx context.Context, identifier string, verifyType string, verifyCode string, action string) error {
	key, err := s.vcKey(verifyType, identifier, action)
	if err != nil {
		return err
	}

	mgr, err := s.vcManager(ctx)
	if err != nil {
		logger.Errorf(ctx, "初始化验证码管理器失败: %v", err)
		return lumerrors.ErrVerifyCodeWrong
	}
	if !mgr.Verify(key, verifyCode) {
		return lumerrors.ErrVerifyCodeWrong
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
