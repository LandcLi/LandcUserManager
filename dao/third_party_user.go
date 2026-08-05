package dao

import (
	"context"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// ThirdPartyUserDao defines the third-party user data access interface.
type ThirdPartyUserDao interface {
	// Create creates a third-party user association
	Create(ctx context.Context, thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error)

	// FindByID finds by ID
	FindByID(ctx context.Context, id uint) (*model.ThirdPartyUserMeta, error)

	// FindByOpenIDAndProvider finds by OpenID and provider
	FindByOpenIDAndProvider(ctx context.Context, openID string, provider string) (*model.ThirdPartyUserMeta, error)

	// FindByUnionID finds by UnionID
	FindByUnionID(ctx context.Context, unionID string) (*model.ThirdPartyUserMeta, error)

	// FindByUserID finds by user ID
	FindByUserID(ctx context.Context, userID uint) ([]*model.ThirdPartyUserMeta, error)

	// Update updates a third-party user association
	Update(ctx context.Context, thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error)

	// Delete deletes a third-party user association
	Delete(ctx context.Context, id uint) error

	// DeleteByUserID deletes by user ID
	DeleteByUserID(ctx context.Context, userID uint) error
}

func RegisterThirdPartyUserDao(impl ThirdPartyUserDao) {
	di.Provide[ThirdPartyUserDao]("thirdPartyUser.dao", impl)
}

func GetThirdPartyUserDao() ThirdPartyUserDao {
	return di.Require[ThirdPartyUserDao]("thirdPartyUser.dao")
}
