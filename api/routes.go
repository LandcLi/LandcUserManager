package api

import (
	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
	"github.com/LandcLi/LandcUserManager/internal/middleware"
	"github.com/gin-gonic/gin"
)

// PublicUserRoutes are user routes that don't require authentication.
var PublicUserRoutes = []string{"Login", "Register", "ResetPassword"}

// ProtectedUserRoutes are user routes that require authentication.
var ProtectedUserRoutes = []string{"LoginRecord", "GetUserInfo", "UpdateUserInfo", "ChangePassword"}

// RegisterDefaultRoutes registers all default routes.
// Public routes (login, register, send-verify-code, reset-password) do not require authentication.
// Protected routes (get-user-info, update-user-info, change-password, login-record) require a valid token.
//
// If no TokenValidator is registered, the Auth middleware operates in no-op mode
// (all requests are allowed), which is useful for local integration where the
// parent system (e.g., SSO) handles authentication at a higher level.
func RegisterDefaultRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", HealthCheck)

	// Common routes (public)
	common.CommonGateway.RegisterRoutes(r)

	// User public routes (no auth required)
	user.UserGateway.RegisterRoutesFiltered(r, PublicUserRoutes)

	// User protected routes (auth required)
	protected := r.Group("")
	protected.Use(middleware.Auth())
	user.UserGateway.RegisterRoutesFiltered(protected, ProtectedUserRoutes)
}

// HealthCheck health check endpoint
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
