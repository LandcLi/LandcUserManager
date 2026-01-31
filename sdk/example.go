package sdk

import (
	"context"
	"fmt"

	"github.com/LandcLi/LandcUserManager/api/common/v1"
	userV1 "github.com/LandcLi/LandcUserManager/api/user/v1"
)

// ExampleSDK 展示SDK的使用方法
func ExampleSDK() {
	// 创建SDK实例
	sdkClient := NewSDK("http://localhost:8080")

	// 1. 基本使用 - 发送验证码
	ctx := context.Background()
	sendVerifyCodeReq := &v1.SendVerifyCodeRequest{
		Identifier: "user@example.com",
		VerifyType: "EMAIL",
		Action:     "LOGIN",
	}
	sendVerifyCodeResp, err := sdkClient.Common.SendVerifyCode(ctx, sendVerifyCodeReq)
	if err != nil {
		fmt.Printf("发送验证码失败: %v\n", err)
		return
	}
	fmt.Printf("发送验证码成功: %v\n", sendVerifyCodeResp.Success)

	// 2. 使用自定义path和header
	customCtx := context.WithValue(context.Background(), "path", "/api/custom/login")
	customCtx = context.WithValue(customCtx, "headers", map[string]string{
		"Authorization": "Bearer token123",
		"X-Custom-Header": "custom-value",
	})

	loginReq := &userV1.LoginRequest{
		Identifier:        "user@example.com",
		AuthCode:          "123456",
		VerifyType:        "EMAIL",
		AuthType:          "VERIFY_CODE",
		IP:                "127.0.0.1",
		DeviceFingerprint: "device123",
		UserAgent:         "Mozilla/5.0",
	}

	loginResp, err := sdkClient.User.Login(customCtx, loginReq)
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}
	fmt.Printf("登录成功，用户ID: %d\n", loginResp.UserID)

	// 3. 获取用户信息
	getUserInfoReq := &userV1.GetUserInfoRequest{
		UserID: loginResp.UserID,
	}
	getUserInfoResp, err := sdkClient.User.GetUserInfo(ctx, getUserInfoReq)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v\n", err)
		return
	}
	fmt.Printf("获取用户信息成功，用户名: %s\n", getUserInfoResp.Username)
}
