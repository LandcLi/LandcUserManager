package middleware

import (
	"strings"

	frameauth "github.com/LandcLi/landc-go/frame/pkg/auth"
	"github.com/LandcLi/landc-go/frame/pkg/response"
	"github.com/gin-gonic/gin"
)

// publicPaths 无需鉴权的公开路径
var publicPaths = map[string]bool{
	"/api/user/login":              true,
	"/api/user/register":           true,
	"/api/user/reset-password":     true,
	"/api/common/send-verify-code": true,
	// 健康检查端点
	"/health": true,
	"/ready":  true,
}

// Auth returns a Gin middleware that validates Bearer tokens using framework JWT.
// 公开路径（登录/注册/重置密码/发送验证码）直接放行，其余路径要求携带有效 token，
// 并将解析出的 UserID 注入 gin.Context，供控制器从 token 取值（防止 IDOR）。
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 公开路径放行
		if publicPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := frameauth.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("UserID", claims.UserID)
		c.Set("Username", claims.Username)
		c.Set("Email", claims.Username)
		c.Set("Scope", claims.Scope)
		c.Set("ClientID", claims.ClientID)

		c.Next()
	}
}
