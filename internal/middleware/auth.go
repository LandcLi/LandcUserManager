package middleware

import (
	"strings"

	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/landc-go/frame/pkg/response"
	"github.com/gin-gonic/gin"
)

// Auth returns a Gin middleware that validates Bearer tokens.
// It uses the registered TokenValidator to validate tokens.
// If no TokenValidator is registered, all requests are allowed (no-op mode).
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		validator := service.GetTokenValidator()
		if validator == nil {
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

		claims, err := validator.ValidateToken(c.Request.Context(), token)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		c.Set("UserID", claims.UserID)
		c.Set("Username", claims.Username)
		c.Set("Email", claims.Email)
		c.Set("Scope", claims.Scope)
		c.Set("ClientID", claims.ClientID)
		c.Set("TokenClaims", claims)

		c.Next()
	}
}
