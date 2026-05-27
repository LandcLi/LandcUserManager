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

	// Register common routes via Gateway
	common.CommonGateway.RegisterRoutes(r)

	// Register user routes via Gateway
	user.UserGateway.RegisterRoutes(r)
}

// HealthCheck health check endpoint
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
