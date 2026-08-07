package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/LandcLi/landc-go/frame/pkg/di"
	framevc "github.com/LandcLi/landc-go/frame/pkg/verifycode"
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

// SendVerifyCode 发送验证码。
// 生成、存储（TTL 5min）、发送间隔（60s）与每日上限（10）由 frame/pkg/verifycode
// （基于 tools/verifycode + tools/ratelimit）承担，本方法只负责发送渠道。
func (s *commonServiceImpl) SendVerifyCode(ctx context.Context, req *model.SendVerifyCodeRequest) error {
	key, err := s.vcKey(req.VerifyType, req.Identifier, req.Action)
	if err != nil {
		return err
	}

	verifyCode, err := framevc.Generate(ctx, key)
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

	switch req.VerifyType {
	case constants.VerifyTypeEmail:
		if err := s.sendMail(ctx, req.Identifier, "验证码", fmt.Sprintf("您的验证码是: %s，5分钟内有效", verifyCode)); err != nil {
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

	if !framevc.Verify(ctx, key, verifyCode) {
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
