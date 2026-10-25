package third_party

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LandcLi/LandcUserManager/model"
	lumerrors "github.com/LandcLi/LandcUserManager/pkg/errors"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/LandcUserManager/utility/logger"
)

// -------------------------- 微信专属Config（内部屏蔽，不对外暴露） --------------------------
type wechatConfig struct {
	AppId     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// -------------------------- 微信适配器实现 --------------------------
type wechatAdapter struct {
	config *wechatConfig
}

func (a *wechatAdapter) FetchUserInfo(ctx context.Context, accessToken, verifyCode string) (*model.ThirdPartyUser, error) {
	// 微信 userinfo 必须携带 openid（access_token 无法反查），verifyCode 即 openid（来自换 token 响应）
	if verifyCode == "" {
		return nil, lumerrors.ErrThirdPartyTokenInvalid
	}
	if accessToken == "" {
		return nil, lumerrors.ErrThirdPartyTokenInvalid
	}
	// 实现微信用户信息拉取逻辑（使用config中的参数）
	// 2. 使用access_token获取用户信息
	userInfoURL := "https://api.weixin.qq.com/sns/userinfo"
	userInfoParams := fmt.Sprintf("access_token=%s&openid=%s", accessToken, verifyCode)

	userInfoResp, err := http.Get(userInfoURL + "?" + userInfoParams)
	if err != nil {
		logger.Errorf(ctx, "获取微信用户信息失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrFetchUserInfoFailed, err)
	}
	defer userInfoResp.Body.Close()

	var userInfoData struct {
		OpenID     string `json:"openid"`
		Nickname   string `json:"nickname"`
		Sex        int    `json:"sex"`
		Province   string `json:"province"`
		City       string `json:"city"`
		Country    string `json:"country"`
		HeadImgURL string `json:"headimgurl"`
		UnionID    string `json:"unionid"`
		Errcode    int    `json:"errcode"`
		Errmsg     string `json:"errmsg"`
	}

	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfoData); err != nil {
		logger.Errorf(ctx, "解析微信用户信息响应失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrParseResponseFailed, err)
	}
	if userInfoData.Errcode != 0 {
		logger.Errorf(ctx, "微信API错误: %d, %s", userInfoData.Errcode, userInfoData.Errmsg)
		return nil, fmt.Errorf("wechat %s%s", lumerrors.ErrAPIError, userInfoData.Errmsg)
	}

	// 转换性别
	gender := ""
	switch userInfoData.Sex {
	case 1:
		gender = "男"
	case 2:
		gender = "女"
	default:
		gender = "未知"
	}

	return &model.ThirdPartyUser{
		AccessToken: accessToken,
		Provider:    "wechat",
		OpenID:      userInfoData.OpenID,
		UnionID:     userInfoData.UnionID,
		Nickname:    userInfoData.Nickname,
		Avatar:      userInfoData.HeadImgURL,
		Email:       "",
		Phone:       "",
		Gender:      gender,
		Country:     userInfoData.Country,
		Province:    userInfoData.Province,
		City:        userInfoData.City,
	}, nil
}

// -------------------------- 微信工厂实现（内部屏蔽Config解析） --------------------------
type wechatFactory struct{}

// Type 返回适配器类型，与配置中的key对应
func (f *wechatFactory) Type() string {
	return "wechat"
}

// Create 解析通用配置为微信专属Config，创建适配器
// 对外屏蔽Config细节，用户只需传map[string]interface{}
func (f *wechatFactory) Create(config map[string]interface{}) (service.ThirdPartyAdapter, error) {
	// 1. 将通用配置转为JSON，再解析为微信专属Config（屏蔽解析细节）
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	var wcConfig wechatConfig
	if err := json.Unmarshal(configBytes, &wcConfig); err != nil {
		return nil, err
	}

	// 2. 校验配置（内部逻辑，不对外暴露）
	if wcConfig.AppId == "" || wcConfig.AppSecret == "" {
		return nil, err
	}

	// 3. 创建适配器实例
	return &wechatAdapter{config: &wcConfig}, nil
}

// -------------------------- 注册默认工厂（库初始化时自动注册） --------------------------
func init() {
	// 注册微信默认工厂，用户可通过RegisterAdapterFactory覆盖
	if _, ok := service.GetFactory("wechat"); !ok {
		service.RegisterAdapterFactory(&wechatFactory{})
	}
}
