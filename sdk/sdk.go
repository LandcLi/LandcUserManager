package sdk

import (
	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
)

// SDK 完整的SDK实例（基于 di.ProxyClient）
type SDK struct {
	User   user.UserController
	Common common.CommonController
}

// NewSDK 创建完整的SDK实例（远程模式，通过 HTTP 代理调用）
func NewSDK(baseURL string) *SDK {
	return &SDK{
		User:   NewRemoteUserController(baseURL),
		Common: NewRemoteCommonController(baseURL),
	}
}
