package third_party

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LandcLi/LandcUserManager/model"
	lumerrors "github.com/LandcLi/LandcUserManager/pkg/errors"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/LandcUserManager/utility/logger"
)

// -------------------------- 钉钉专属Config（内部屏蔽，不对外暴露） --------------------------
type dingtalkConfig struct {
	AppId     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// -------------------------- 钉钉适配器实现 --------------------------
type dingtalkAdapter struct {
	config *dingtalkConfig
}

func (a *dingtalkAdapter) FetchUserInfo(ctx context.Context, accessToken, verifyCode string) (*model.ThirdPartyUser, error) {
	// 实现钉钉用户信息拉取逻辑（使用config中的参数）
	// 2. 使用access_token和code获取用户信息
	userInfoURL := "https://oapi.dingtalk.com/sns/getuserinfo_bycode"

	// 构建请求体
	reqBody := map[string]string{
		"tmp_auth_code": verifyCode,
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		logger.Errorf(ctx, "构建钉钉用户信息请求失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrBuildRequestFailed, err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", userInfoURL, strings.NewReader(string(reqJSON)))
	if err != nil {
		logger.Errorf(ctx, "创建钉钉用户信息请求失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrCreateRequestFailed, err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", accessToken)

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	userInfoResp, err := client.Do(req)
	if err != nil {
		logger.Errorf(ctx, "获取钉钉用户信息失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrFetchUserInfoFailed, err)
	}
	defer userInfoResp.Body.Close()

	var userInfoData struct {
		Errcode  int    `json:"errcode"`
		Errmsg   string `json:"errmsg"`
		UserInfo struct {
			OpenID  string `json:"openid"`
			UnionID string `json:"unionid"`
			Nick    string `json:"nick"`
			Avatar  string `json:"avatar"`
		} `json:"user_info"`
	}

	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfoData); err != nil {
		logger.Errorf(ctx, "解析钉钉用户信息响应失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrParseResponseFailed, err)
	}

	if userInfoData.Errcode != 0 {
		logger.Errorf(ctx, "钉钉API错误: %d, %s", userInfoData.Errcode, userInfoData.Errmsg)
		return nil, fmt.Errorf("dingtalk %s%s", lumerrors.ErrAPIError, userInfoData.Errmsg)
	}

	return &model.ThirdPartyUser{
		Provider: "dingtalk",
		OpenID:   userInfoData.UserInfo.OpenID,
		UnionID:  userInfoData.UserInfo.UnionID,
		Nickname: userInfoData.UserInfo.Nick,
		Avatar:   userInfoData.UserInfo.Avatar,
		Email:    "",
		Phone:    "",
		Gender:   "",
		Country:  "",
		Province: "",
		City:     "",
	}, nil
}

// -------------------------- 钉钉工厂实现（内部屏蔽Config解析） --------------------------
type dingtalkFactory struct{}

// Type 返回适配器类型，与配置中的key对应
func (f *dingtalkFactory) Type() string {
	return "dingtalk"
}

// Create 解析通用配置为钉钉专属Config，创建适配器
// 对外屏蔽Config细节，用户只需传map[string]interface{}
func (f *dingtalkFactory) Create(config map[string]interface{}) (service.ThirdPartyAdapter, error) {
	// 1. 将通用配置转为JSON，再解析为钉钉专属Config（屏蔽解析细节）
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	var dcConfig dingtalkConfig
	if err := json.Unmarshal(configBytes, &dcConfig); err != nil {
		return nil, err
	}

	// 2. 校验配置（内部逻辑，不对外暴露）
	if dcConfig.AppId == "" || dcConfig.AppSecret == "" {
		return nil, err
	}

	// 3. 创建适配器实例
	return &dingtalkAdapter{config: &dcConfig}, nil
}

// -------------------------- 注册默认工厂（库初始化时自动注册） --------------------------
func init() {
	// 注册钉钉默认工厂，用户可通过RegisterAdapterFactory覆盖
	if _, ok := service.GetFactory("dingtalk"); !ok {
		service.RegisterAdapterFactory(&dingtalkFactory{})
	}
}
