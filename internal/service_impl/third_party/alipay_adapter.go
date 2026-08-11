package third_party

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/LandcLi/LandcUserManager/model"
	lumerrors "github.com/LandcLi/LandcUserManager/pkg/errors"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/LandcUserManager/utility/logger"
)

// -------------------------- 支付宝专属Config（内部屏蔽，不对外暴露） --------------------------
type alipayConfig struct {
	AppId     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// -------------------------- 支付宝适配器实现 --------------------------
type alipayAdapter struct {
	config *alipayConfig
}

func (a *alipayAdapter) FetchUserInfo(ctx context.Context, accessToken, verifyCode string) (*model.ThirdPartyUser, error) {
	// 实现支付宝用户信息拉取逻辑（使用config中的参数）
	// 1. 使用access_token获取用户信息
	userInfoURL := "https://openapi.alipay.com/gateway.do"

	// 构建请求参数
	userInfoParams := map[string]string{
		"app_id":     a.config.AppId,
		"method":     "alipay.user.info.share",
		"format":     "JSON",
		"charset":    "UTF-8",
		"sign_type":  "RSA2",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"auth_token": accessToken,
	}

	// TODO: implement Alipay signature logic (planned feature)
	// 实际项目中，需要根据支付宝文档生成签名

	// 构建请求URL
	userInfoParamsStr := ""
	for k, v := range userInfoParams {
		userInfoParamsStr += k + "=" + url.QueryEscape(v) + "&"
	}
	userInfoParamsStr = userInfoParamsStr[:len(userInfoParamsStr)-1]

	userInfoResp, err := http.Get(userInfoURL + "?" + userInfoParamsStr)
	if err != nil {
		logger.Errorf(ctx, "获取支付宝用户信息失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrFetchUserInfoFailed, err)
	}
	defer userInfoResp.Body.Close()

	var userInfoData struct {
		AlipayUserInfoShareResponse struct {
			UserId   string `json:"user_id"`
			Avatar   string `json:"avatar"`
			City     string `json:"city"`
			Province string `json:"province"`
			NickName string `json:"nick_name"`
			Gender   string `json:"gender"`
			Email    string `json:"email"`
		} `json:"alipay_user_info_share_response"`
		ErrorResponse struct {
			Code    string `json:"code"`
			Msg     string `json:"msg"`
			SubCode string `json:"sub_code"`
			SubMsg  string `json:"sub_msg"`
		} `json:"error_response"`
	}

	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfoData); err != nil {
		logger.Errorf(ctx, "解析支付宝用户信息响应失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrParseResponseFailed, err)
	}

	// 检查是否有错误
	if userInfoData.ErrorResponse.Code != "" {
		logger.Errorf(ctx, "支付宝API错误: %s, %s", userInfoData.ErrorResponse.SubCode, userInfoData.ErrorResponse.SubMsg)
		return nil, fmt.Errorf("alipay %s%s", lumerrors.ErrAPIError, userInfoData.ErrorResponse.SubMsg)
	}

	return &model.ThirdPartyUser{
		AccessToken: accessToken,
		Provider: "alipay",
		OpenID:   userInfoData.AlipayUserInfoShareResponse.UserId,
		UnionID:  userInfoData.AlipayUserInfoShareResponse.UserId,
		Nickname: userInfoData.AlipayUserInfoShareResponse.NickName,
		Avatar:   userInfoData.AlipayUserInfoShareResponse.Avatar,
		Email:    userInfoData.AlipayUserInfoShareResponse.Email,
		Phone:    "",
		Gender:   userInfoData.AlipayUserInfoShareResponse.Gender,
		Country:  "",
		Province: userInfoData.AlipayUserInfoShareResponse.Province,
		City:     userInfoData.AlipayUserInfoShareResponse.City,
	}, nil
}

// -------------------------- 支付宝工厂实现（内部屏蔽Config解析） --------------------------
type alipayFactory struct{}

// Type 返回适配器类型，与配置中的key对应
func (f *alipayFactory) Type() string {
	return "alipay"
}

// Create 解析通用配置为支付宝专属Config，创建适配器
// 对外屏蔽Config细节，用户只需传map[string]interface{}
func (f *alipayFactory) Create(config map[string]interface{}) (service.ThirdPartyAdapter, error) {
	// 1. 将通用配置转为JSON，再解析为支付宝专属Config（屏蔽解析细节）
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	var acConfig alipayConfig
	if err := json.Unmarshal(configBytes, &acConfig); err != nil {
		return nil, err
	}

	// 2. 校验配置（内部逻辑，不对外暴露）
	if acConfig.AppId == "" || acConfig.AppSecret == "" {
		return nil, err
	}

	// 3. 创建适配器实例
	return &alipayAdapter{config: &acConfig}, nil
}

// -------------------------- 注册默认工厂（库初始化时自动注册） --------------------------
func init() {
	// 注册支付宝默认工厂，用户可通过RegisterAdapterFactory覆盖
	if _, ok := service.GetFactory("alipay"); !ok {
		service.RegisterAdapterFactory(&alipayFactory{})
	}
}
