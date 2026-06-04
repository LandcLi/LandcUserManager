package middleware

import (
	apiMiddleware "github.com/LandcLi/landc-go/api/middleware/gin"
	"github.com/gin-gonic/gin"
)

// GinMiddleware LandcCommonApi中间件
func GinMiddleware() gin.HandlerFunc {
	return apiMiddleware.GinMiddleware()
}
