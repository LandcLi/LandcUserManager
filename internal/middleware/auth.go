package middleware

import (
	"context"
	"crypto/subtle"
	"strings"

	frameauth "github.com/LandcLi/landc-go/frame/pkg/auth"
	"github.com/LandcLi/landc-go/frame/pkg/response"
	"github.com/gin-gonic/gin"

	"github.com/LandcLi/LandcUserManager/utility/config"
)

// ctxKey 用于在 request context 中传递鉴权信息。
// 注意：landc-go createHandler 传给 controller 的是 c.Request.Context()（标准 Go context），
// 因此鉴权信息必须注入 request context，而非仅存 gin.Context keys。
type ctxKey string

const (
	userIDKey   ctxKey = "auth.user_id"
	usernameKey ctxKey = "auth.username"
	scopeKey    ctxKey = "auth.scope"
	clientIDKey ctxKey = "auth.client_id"
)

// UserIDFromContext 从 request context 读取当前登录用户 ID（由 Auth 中间件注入）。
func UserIDFromContext(ctx context.Context) (uint, bool) {
	v, ok := ctx.Value(userIDKey).(uint)
	return v, ok
}

// UsernameFromContext 从 request context 读取当前登录用户名。
func UsernameFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(usernameKey).(string)
	return v, ok
}

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

// ServiceKeyHeader 服务间调用凭证的请求头名。
const ServiceKeyHeader = "X-Service-Key"

// isServiceCall 判断请求是否来自可信的内部服务。
//
// 用途：LandcSSO 的 OIDC /userinfo 要按令牌里的 sub 读取用户档案，这是服务间调用，
// 手上没有「当前登录用户」的 token。共享密钥给这条路留口子，外部客户端无法伪造。
//
// 未配置密钥（cfg.ServiceAuth.Key == ""）时恒为 false —— 漏配只会更严格，不会放开。
func isServiceCall(c *gin.Context) bool {
	cfg := config.TryGetConfig()
	if cfg == nil || cfg.ServiceAuth.Key == "" {
		return false
	}
	got := c.GetHeader(ServiceKeyHeader)
	if got == "" {
		return false
	}
	// 恒定时间比较，避免通过响应耗时逐字节猜密钥
	return subtle.ConstantTimeCompare([]byte(got), []byte(cfg.ServiceAuth.Key)) == 1
}

// Auth returns a Gin middleware that validates Bearer tokens using framework JWT.
// 公开路径（登录/注册/重置密码/发送验证码）直接放行，其余路径要求携带有效 token，
// 并将解析出的 claims 注入 request context 与 gin.Context，供控制器从 token 取值（防止 IDOR）。
//
// 服务间调用（携带正确的 X-Service-Key）也放行，用于 SSO 按 user_id 反查用户档案。
// 这类请求没有用户身份，因此不注入 user_id —— 控制器必须显式使用请求体里的 user_id。
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 公开路径放行
		if publicPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// 服务间调用放行（不注入用户身份）
		if isServiceCall(c) {
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

		// 注入 request context（createHandler 传 c.Request.Context() 给 controller）
		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, userIDKey, claims.UserID)
		reqCtx = context.WithValue(reqCtx, usernameKey, claims.Username)
		reqCtx = context.WithValue(reqCtx, scopeKey, claims.Scope)
		reqCtx = context.WithValue(reqCtx, clientIDKey, claims.ClientID)
		c.Request = c.Request.WithContext(reqCtx)

		// 保留 gin.Context keys（兼容 gin 原生 handler 用法）
		c.Set("UserID", claims.UserID)
		c.Set("Username", claims.Username)
		c.Set("Email", claims.Username)
		c.Set("Scope", claims.Scope)
		c.Set("ClientID", claims.ClientID)

		c.Next()
	}
}
