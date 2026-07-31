package v1

import (
	"time"

	"github.com/LandcLi/landc-go/frame/pkg/meta"
)

// UpdateUserInfoRequest 更新用户信息请求
// 注意：UserID 不从前端获取，而是由 Auth 中间件从 JWT claims 中解析并注入。
type UpdateUserInfoRequest struct {
	meta.Meta `path:"/api/user/update-user-info" method:"POST"`
	Nickname  string `json:"nickname" binding:"omitempty,max=50"`
	Phone     string `json:"phone" binding:"omitempty,max=20"`
	Avatar    string `json:"avatar" binding:"omitempty,max=255"`
}

// UpdateUserInfoResponse 更新用户信息响应
type UpdateUserInfoResponse struct {
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
