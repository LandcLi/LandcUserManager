package sdk

import (
	"context"

	commonV1 "github.com/LandcLi/LandcUserManager/api/common/v1"
	userV1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// ========================== RemoteUserController ==========================

// RemoteUserController implements user.UserController via HTTP proxy.
type RemoteUserController struct {
	client *di.ProxyClient
}

// NewRemoteUserController creates a remote user controller.
//
//	remote := sdk.NewRemoteUserController("http://user-service:8081")
//	user.UserGateway.Override(remote)
func NewRemoteUserController(baseURL string, opts ...di.RemoteOption) *RemoteUserController {
	return &RemoteUserController{
		client: di.NewProxyClient(baseURL, opts...),
	}
}

func (r *RemoteUserController) Login(ctx context.Context, req *userV1.LoginRequest) (*userV1.LoginResponse, error) {
	return di.Call[userV1.LoginResponse](r.client, ctx, "Login", req)
}

func (r *RemoteUserController) Register(ctx context.Context, req *userV1.RegisterRequest) (*userV1.RegisterResponse, error) {
	return di.Call[userV1.RegisterResponse](r.client, ctx, "Register", req)
}

func (r *RemoteUserController) LoginRecord(ctx context.Context, req *userV1.LoginRecordRequest) (*userV1.LoginRecordResponse, error) {
	return di.Call[userV1.LoginRecordResponse](r.client, ctx, "LoginRecord", req)
}

func (r *RemoteUserController) GetUserInfo(ctx context.Context, req *userV1.GetUserInfoRequest) (*userV1.GetUserInfoResponse, error) {
	return di.Call[userV1.GetUserInfoResponse](r.client, ctx, "GetUserInfo", req)
}

func (r *RemoteUserController) UpdateUserInfo(ctx context.Context, req *userV1.UpdateUserInfoRequest) (*userV1.UpdateUserInfoResponse, error) {
	return di.Call[userV1.UpdateUserInfoResponse](r.client, ctx, "UpdateUserInfo", req)
}

func (r *RemoteUserController) ChangePassword(ctx context.Context, req *userV1.ChangePasswordRequest) (*userV1.ChangePasswordResponse, error) {
	return di.Call[userV1.ChangePasswordResponse](r.client, ctx, "ChangePassword", req)
}

func (r *RemoteUserController) ResetPassword(ctx context.Context, req *userV1.ResetPasswordRequest) (*userV1.ResetPasswordResponse, error) {
	return di.Call[userV1.ResetPasswordResponse](r.client, ctx, "ResetPassword", req)
}

// ========================== RemoteCommonController ==========================

// RemoteCommonController implements common.CommonController via HTTP proxy.
type RemoteCommonController struct {
	client *di.ProxyClient
}

// NewRemoteCommonController creates a remote common controller.
func NewRemoteCommonController(baseURL string, opts ...di.RemoteOption) *RemoteCommonController {
	return &RemoteCommonController{
		client: di.NewProxyClient(baseURL, opts...),
	}
}

func (r *RemoteCommonController) SendVerifyCode(ctx context.Context, req *commonV1.SendVerifyCodeRequest) (*commonV1.SendVerifyCodeResponse, error) {
	return di.Call[commonV1.SendVerifyCodeResponse](r.client, ctx, "SendVerifyCode", req)
}
