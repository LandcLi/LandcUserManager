package sdk

import (
	"context"

	v1 "github.com/LandcLi/LandcUserManager/api/common/v1"
)

// Common 通用相关接口
type Common struct {
	client *Client
}

// NewCommon 创建通用相关接口实例
func NewCommon(client *Client) *Common {
	return &Common{client: client}
}

// SendVerifyCode 发送验证码
func (c *Common) SendVerifyCode(ctx context.Context, req *v1.SendVerifyCodeRequest) (*v1.SendVerifyCodeResponse, error) {
	resp := &v1.SendVerifyCodeResponse{}
	err := c.client.post(ctx, "/api/common/send-verify-code", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
