package user

import (
	"context"
	"net/http"

	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/gin-gonic/gin"
)

var controller userController

type userController interface {
	Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error)
}

func RegisterUserController(impl userController) {
	controller = impl
}

func GetUserController() userController {
	return controller
}

func Login(c *gin.Context) {
	var req v1.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := controller.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
