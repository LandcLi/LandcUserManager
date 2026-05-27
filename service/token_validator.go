package service

import (
	"context"

	"github.com/LandcLi/landc-go/mvc/pkg/di"
)

// TokenClaims represents the validated token claims.
// This is a simplified version that UserManager needs from the token.
type TokenClaims struct {
	UserID            uint   `json:"user_id"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	Scope             string `json:"scope"`
	ClientID          string `json:"client_id"`
	IP                string `json:"ip"`
	DeviceFingerprint string `json:"device_fingerprint"`
}

// TokenValidator defines the token validation interface.
// SSO maintains login state, UserManager uses this interface to validate tokens.
//
// Local mode (SSO integrated): directly calls SSO's authService.ValidateToken()
// Remote mode (standalone): calls SSO's /auth/introspect endpoint via HTTP
//
// Users can implement this interface and register to customize token validation.
//
// Example:
//
//	// Local mode - SSO provides implementation
//	service.RegisterTokenValidator(ssoLocalValidator)
//
//	// Remote mode - call SSO via HTTP
//	service.RegisterTokenValidator(sdk.NewRemoteTokenValidator("http://sso:8080"))
//
//	// Custom mode - use your own JWT validation
//	service.RegisterTokenValidator(myJWTValidator)
type TokenValidator interface {
	// ValidateToken validates the token and returns claims if valid.
	// Returns error if token is invalid, expired, or revoked.
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

func RegisterTokenValidator(impl TokenValidator) {
	di.Provide[TokenValidator]("token.validator", impl)
}

func GetTokenValidator() TokenValidator {
	v, err := di.Resolve[TokenValidator]("token.validator")
	if err != nil {
		return nil
	}
	return v
}
