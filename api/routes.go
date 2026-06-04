package api

import (
	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
	"github.com/LandcLi/LandcUserManager/internal/middleware"
	"github.com/LandcLi/landc-go/frame/pkg/response"
	"github.com/gin-gonic/gin"
)

// PublicUserRoutes are user routes that don't require authentication.
var PublicUserRoutes = []string{"Login", "Register", "ResetPassword"}

// ProtectedUserRoutes are user routes that require authentication.
var ProtectedUserRoutes = []string{"LoginRecord", "GetUserInfo", "UpdateUserInfo", "ChangePassword"}

// RegisterDefaultRoutes registers all default routes on any gin.IRouter.
func RegisterDefaultRoutes(r gin.IRouter) {
	// Health check
	if eg, ok := r.(*gin.Engine); ok {
		eg.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"})
		})
	} else if rg, ok := r.(*gin.RouterGroup); ok {
		rg.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"})
		})
	}

	// Common routes (public)
	common.CommonGateway.RegisterRoutes(r)

	// User public routes (no auth required)
	user.UserGateway.RegisterRoutesFiltered(r, PublicUserRoutes)

	// User protected routes (auth required)
	if eg, ok := r.(*gin.Engine); ok {
		protected := eg.Group("")
		protected.Use(middleware.Auth())
		user.UserGateway.RegisterRoutesFiltered(protected, ProtectedUserRoutes)
	} else if rg, ok := r.(*gin.RouterGroup); ok {
		protected := rg.Group("")
		protected.Use(middleware.Auth())
		user.UserGateway.RegisterRoutesFiltered(protected, ProtectedUserRoutes)
	}
}
