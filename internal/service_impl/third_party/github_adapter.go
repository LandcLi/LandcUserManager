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

// -------------------------- GitHub专属Config（内部屏蔽，不对外暴露） --------------------------
type githubConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// -------------------------- GitHub适配器实现 --------------------------
type githubAdapter struct {
	config *githubConfig
}

func (a *githubAdapter) FetchUserInfo(ctx context.Context, accessToken, verifyCode string) (*model.ThirdPartyUser, error) {
	// 实现GitHub用户信息拉取逻辑（使用config中的参数）
	client := &http.Client{}
	// 2. 使用access_token获取用户信息
	userInfoURL := "https://api.github.com/user"

	// 创建请求
	userInfoReq, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		logger.Errorf(ctx, "创建GitHub用户信息请求失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrCreateRequestFailed, err)
	}

	// 设置请求头
	userInfoReq.Header.Set("Authorization", "Bearer "+accessToken)
	userInfoReq.Header.Set("Accept", "application/vnd.github.v3+json")

	// 发送请求
	userInfoResp, err := client.Do(userInfoReq)
	if err != nil {
		logger.Errorf(ctx, "获取GitHub用户信息失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrFetchUserInfoFailed, err)
	}
	defer userInfoResp.Body.Close()

	var userInfoData struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Bio       string `json:"bio"`
		Location  string `json:"location"`
		Error     string `json:"error"`
	}

	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfoData); err != nil {
		logger.Errorf(ctx, "解析GitHub用户信息响应失败: %v", err)
		return nil, fmt.Errorf("%s: %w", lumerrors.ErrParseResponseFailed, err)
	}

	if userInfoData.Error != "" {
		logger.Errorf(ctx, "GitHub API错误: %s", userInfoData.Error)
		return nil, fmt.Errorf("github %s%s", lumerrors.ErrAPIError, userInfoData.Error)
	}

	// 3. 如果email为空，尝试获取用户的公开邮箱
	if userInfoData.Email == "" {
		emailsURL := "https://api.github.com/user/emails"
		emailsReq, err := http.NewRequest("GET", emailsURL, nil)
		if err == nil {
			emailsReq.Header.Set("Authorization", "Bearer "+accessToken)
			emailsReq.Header.Set("Accept", "application/vnd.github.v3+json")

			emailsResp, err := client.Do(emailsReq)
			if err == nil {
				defer emailsResp.Body.Close()

				var emailsData []struct {
					Email    string `json:"email"`
					Primary  bool   `json:"primary"`
					Verified bool   `json:"verified"`
				}

				if json.NewDecoder(emailsResp.Body).Decode(&emailsData) == nil {
					for _, email := range emailsData {
						if email.Primary && email.Verified {
							userInfoData.Email = email.Email
							break
						}
					}
				}
			}
		}
	}

	nickname := userInfoData.Name
	if nickname == "" {
		nickname = userInfoData.Login
	}

	return &model.ThirdPartyUser{
		AccessToken: accessToken,
		Provider: "github",
		OpenID:   fmt.Sprintf("%d", userInfoData.ID),
		UnionID:  fmt.Sprintf("%d", userInfoData.ID),
		Nickname: nickname,
		Avatar:   userInfoData.AvatarURL,
		Email:    userInfoData.Email,
		Phone:    "",
		Gender:   "",
		Country:  "",
		Province: "",
		City:     userInfoData.Location,
	}, nil
}

// -------------------------- GitHub工厂实现（内部屏蔽Config解析） --------------------------
type githubFactory struct{}

// Type 返回适配器类型，与配置中的key对应
func (f *githubFactory) Type() string {
	return "github"
}

// Create 解析通用配置为GitHub专属Config，创建适配器
// 对外屏蔽Config细节，用户只需传map[string]interface{}
func (f *githubFactory) Create(config map[string]interface{}) (service.ThirdPartyAdapter, error) {
	// 1. 将通用配置转为JSON，再解析为GitHub专属Config（屏蔽解析细节）
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	var gcConfig githubConfig
	if err := json.Unmarshal(configBytes, &gcConfig); err != nil {
		return nil, err
	}

	// 2. 校验配置（内部逻辑，不对外暴露）
	if gcConfig.ClientId == "" || gcConfig.ClientSecret == "" {
		return nil, err
	}

	// 3. 创建适配器实例
	return &githubAdapter{config: &gcConfig}, nil
}

// -------------------------- 注册默认工厂（库初始化时自动注册） --------------------------
func init() {
	// 注册GitHub默认工厂，用户可通过RegisterAdapterFactory覆盖
	if _, ok := service.GetFactory("github"); !ok {
		service.RegisterAdapterFactory(&githubFactory{})
	}
}
