package api

import (
	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
	"github.com/gin-gonic/gin"
)

// RegisterDefaultRoutes registers all default routes using Gateway auto-generation.
func RegisterDefaultRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", HealthCheck)

	// Register common routes
	commonGroup := r.Group("/api/common")
	{
		commonGroup.POST("/send-verify-code", common.SendVerifyCode)
	}

	// Register user routes via Gateway (auto-generates Gin handlers from UserController interface)
	user.UserGateway.RegisterRoutes(r)
}

// HealthCheck health check endpoint
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
