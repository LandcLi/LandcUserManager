package middleware

import (
	"net/http"
	"strings"

	"github.com/LandcLi/LandcUserManager/service"
	"github.com/gin-gonic/gin"
)

// Auth returns a Gin middleware that validates Bearer tokens.
// It uses the registered TokenValidator to validate tokens.
// If no TokenValidator is registered, all requests are allowed (no-op mode).
//
// Usage in routes:
//
//	authorized := r.Group("/api/user")
//	authorized.Use(middleware.Auth())
//	{
//	    authorized.POST("/get-user-info", ...)
//	}
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		validator := service.GetTokenValidator()
		if validator == nil {
			// No validator registered, skip authentication (no-op mode)
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		token := parts[1]

		// Validate token
		claims, err := validator.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Set claims in context for downstream handlers
		c.Set("UserID", claims.UserID)
		c.Set("Username", claims.Username)
		c.Set("Email", claims.Email)
		c.Set("Scope", claims.Scope)
		c.Set("ClientID", claims.ClientID)
		c.Set("TokenClaims", claims)

		c.Next()
	}
}
