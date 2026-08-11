package service

import (
	"context"
	"fmt"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// ThirdPartyAdapter 第三方适配器核心接口
type ThirdPartyAdapter interface {
	FetchUserInfo(ctx context.Context, accessToken, verifyCode string) (*model.ThirdPartyUser, error)
}

// AdapterFactory 适配器工厂接口
type AdapterFactory interface {
	Type() string
	Create(config map[string]interface{}) (ThirdPartyAdapter, error)
}

// RegisterAdapterFactory 注册适配器工厂到 DI 容器
func RegisterAdapterFactory(factory AdapterFactory) {
	di.Override[AdapterFactory](fmt.Sprintf("adapter.factory.%s", factory.Type()), factory)
}

// GetFactory 获取适配器工厂
func GetFactory(adapterType string) (AdapterFactory, bool) {
	f, err := di.Resolve[AdapterFactory](fmt.Sprintf("adapter.factory.%s", adapterType))
	return f, err == nil
}

// RegisterThirdPartyAdapter 注册适配器实例到 DI 容器
func RegisterThirdPartyAdapter(adapterType string, adapter ThirdPartyAdapter) {
	di.Override[ThirdPartyAdapter](fmt.Sprintf("adapter.%s", adapterType), adapter)
}

// GetThirdPartyAdapter 获取适配器实例
func GetThirdPartyAdapter(adapterType string) (ThirdPartyAdapter, bool) {
	a, err := di.Resolve[ThirdPartyAdapter](fmt.Sprintf("adapter.%s", adapterType))
	return a, err == nil
}

// InitWithConfig 库对外暴露的核心初始化函数
// 遍历配置中的所有第三方类型，通过工厂创建并注册适配器
func InitWithConfig(globalConfig *config.ThirdPartyGlobalConfig) error {
	for adapterType, cfgMap := range *globalConfig {
		factory, ok := GetFactory(adapterType)
		if !ok {
			return fmt.Errorf("factory not registered for type %s", adapterType)
		}

		adapter, err := factory.Create(cfgMap)
		if err != nil {
			return fmt.Errorf("failed to create adapter %s: %w", adapterType, err)
		}

		// 运行时动态注册，使用 Override 而非 Provide（Provide 仅在 init 阶段生效）
		RegisterThirdPartyAdapter(adapterType, adapter)
	}
	return nil
}
