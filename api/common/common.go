package common

import (
	"context"
	"net/http"

	v1 "github.com/LandcLi/LandcUserManager/api/common/v1"
	"github.com/gin-gonic/gin"
)

var controller commonController

type commonController interface {
	SendVerifyCode(ctx context.Context, req *v1.SendVerifyCodeRequest) (*v1.SendVerifyCodeResponse, error)
}

func RegisterCommonController(impl commonController) {
	controller = impl
}

func GetCommonController() commonController {
	return controller
}

func SendVerifyCode(c *gin.Context) {
	var req v1.SendVerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := controller.SendVerifyCode(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
