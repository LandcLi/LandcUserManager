package v1

import (
	"time"

	"github.com/LandcLi/landc-go/frame/pkg/meta"
)

// GetUserInfoRequest 获取用户信息请求
type GetUserInfoRequest struct {
	meta.Meta `path:"/api/user/get-user-info" method:"POST"`
	UserID uint `json:"userId" binding:"required"`
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
