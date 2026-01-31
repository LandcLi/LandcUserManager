package api

import (
	"github.com/LandcLi/LandcUserManager/api/common"
	"github.com/LandcLi/LandcUserManager/api/user"
	"github.com/gin-gonic/gin"
)

// RegisterDefaultRoutes 注册默认路由
func RegisterDefaultRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", HealthCheck)

	// 注册common路由
	commonGroup := r.Group("/api/common")
	{
		commonGroup.POST("/send-verify-code", common.SendVerifyCode)
	}

	// 注册user路由
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/login", user.Login)
		userGroup.POST("/register", user.Register)
		userGroup.POST("/login-record", user.LoginRecord)
		userGroup.POST("/get-user-info", user.GetUserInfo)
		userGroup.POST("/update-user-info", user.UpdateUserInfo)
		userGroup.POST("/change-password", user.ChangePassword)
		userGroup.POST("/reset-password", user.ResetPassword)
	}
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
