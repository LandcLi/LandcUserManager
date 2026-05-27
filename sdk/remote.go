package sdk

import (
	"context"

	"github.com/LandcLi/LandcUserManager/api/user"
	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/landc-go/mvc/pkg/di"
)

// RemoteUserController implements user.UserController via HTTP calls.
// Use this when LandcUserManager is deployed as a separate service.
type RemoteUserController struct {
	client *di.ProxyClient
}

// NewRemoteUserController creates a remote user controller that calls
// the user service via HTTP.
//
//	// In main.go to switch to remote mode:
//	remote := sdk.NewRemoteUserController("http://user-service:8081")
//	user.UserGateway.Override(remote)
func NewRemoteUserController(baseURL string, opts ...di.RemoteOption) *RemoteUserController {
	return &RemoteUserController{
		client: di.NewProxyClient(baseURL, user.UserRoutes, opts...),
	}
}

func (r *RemoteUserController) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	return di.Call[v1.LoginResponse](r.client, ctx, "Login", req)
}

func (r *RemoteUserController) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	return di.Call[v1.RegisterResponse](r.client, ctx, "Register", req)
}

func (r *RemoteUserController) LoginRecord(ctx context.Context, req *v1.LoginRecordRequest) (*v1.LoginRecordResponse, error) {
	return di.Call[v1.LoginRecordResponse](r.client, ctx, "LoginRecord", req)
}

func (r *RemoteUserController) GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.GetUserInfoResponse, error) {
	return di.Call[v1.GetUserInfoResponse](r.client, ctx, "GetUserInfo", req)
}

func (r *RemoteUserController) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoResponse, error) {
	return di.Call[v1.UpdateUserInfoResponse](r.client, ctx, "UpdateUserInfo", req)
}

func (r *RemoteUserController) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error) {
	return di.Call[v1.ChangePasswordResponse](r.client, ctx, "ChangePassword", req)
}

func (r *RemoteUserController) ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) (*v1.ResetPasswordResponse, error) {
	return di.Call[v1.ResetPasswordResponse](r.client, ctx, "ResetPassword", req)
}
