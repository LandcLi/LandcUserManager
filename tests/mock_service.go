package tests

import (
	"context"
	"errors"
	"sync"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// MockCommonService mocks the CommonService interface
type MockCommonService struct {
	mu         sync.Mutex
	VerifyCode string // the code to accept; empty = accept none
}

func NewMockCommonService() *MockCommonService {
	return &MockCommonService{}
}

func (m *MockCommonService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.VerifyCode = ""
}

func (m *MockCommonService) SendVerifyCode(ctx context.Context, req *model.SendVerifyCodeRequest) error {
	return nil
}

func (m *MockCommonService) CheckVerifyCode(ctx context.Context, identifier string, verifyType string, verifyCode string, action string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.VerifyCode != "" && verifyCode == m.VerifyCode {
		return nil
	}
	return errors.New("验证码错误")
}

// MockThirdPartyAdapter mocks the ThirdPartyAdapter interface
type MockThirdPartyAdapter struct {
	User *model.ThirdPartyUser
}

func NewMockThirdPartyAdapter(user *model.ThirdPartyUser) *MockThirdPartyAdapter {
	return &MockThirdPartyAdapter{User: user}
}

func (m *MockThirdPartyAdapter) FetchUserInfo(ctx context.Context, accessToken, verifyCode string) (*model.ThirdPartyUser, error) {
	if m.User == nil {
		return nil, errors.New("mock adapter: no user configured")
	}
	return m.User, nil
}

// RegisterMockThirdPartyAdapter registers a mock third-party adapter for testing
func RegisterMockThirdPartyAdapter(provider string, adapter *MockThirdPartyAdapter) {
	di.Override[service.ThirdPartyAdapter]("adapter."+provider, adapter)
}
