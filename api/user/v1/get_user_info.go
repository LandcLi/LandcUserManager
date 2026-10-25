package v1

import (
	"time"

	"github.com/LandcLi/landc-go/frame/pkg/meta"
)

// GetUserInfoRequest 获取用户信息请求。
//
// UserID 的取值优先级：
//  1. 请求里显式指定（服务间调用，例如 LandcSSO 的 OIDC /userinfo 需要按
//     令牌里的 sub 取用户档案，此时没有「当前登录用户」这个概念）；
//  2. 未指定时取 Auth 中间件从 JWT claims 解析并注入的当前用户。
//
// 普通前端调用不要传 user_id —— 那等于把「查谁」的决定权交给客户端。
type GetUserInfoRequest struct {
	meta.Meta `path:"/api/user/get-user-info" method:"POST"`

	// UserID 可选。仅服务间调用使用；0 表示取当前登录用户。
	UserID uint `json:"user_id"`
}

// GetUserInfoResponse 获取用户信息响应
type GetUserInfoResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Phone     string    `json:"phone"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
