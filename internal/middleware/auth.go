package middleware

import (
	"strings"

	frameauth "github.com/LandcLi/landc-go/frame/pkg/auth"
	"github.com/LandcLi/landc-go/frame/pkg/response"
	"github.com/gin-gonic/gin"
)

// Auth returns a Gin middleware that validates Bearer tokens using framework JWT.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
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
