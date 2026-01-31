package sdk

import (
	"net/http"
)

// SDK 完整的SDK实例
type SDK struct {
	Client *Client
	User   *User
	Common *Common
}

// NewSDK 创建完整的SDK实例
func NewSDK(baseURL string) *SDK {
	client := NewClient(baseURL)
	return &SDK{
		Client: client,
		User:   NewUser(client),
		Common: NewCommon(client),
	}
}

// NewSDKWithHTTPClient 使用自定义的HTTP客户端创建SDK实例
func NewSDKWithHTTPClient(baseURL string, httpClient *http.Client) *SDK {
	client := NewClient(baseURL)
	client.HTTPClient = httpClient
	return &SDK{
		Client: client,
		User:   NewUser(client),
		Common: NewCommon(client),
	}
}
