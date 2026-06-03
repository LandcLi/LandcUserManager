package dao

import (
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// ThirdPartyUserDao defines the third-party user data access interface.
type ThirdPartyUserDao interface {
	// Create creates a third-party user association
	Create(thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error)

	// FindByID finds by ID
	FindByID(id uint) (*model.ThirdPartyUserMeta, error)

	// FindByOpenIDAndProvider finds by OpenID and provider
	FindByOpenIDAndProvider(openID string, provider string) (*model.ThirdPartyUserMeta, error)

	// FindByUnionID finds by UnionID
	FindByUnionID(unionID string) (*model.ThirdPartyUserMeta, error)

	// FindByUserID finds by user ID
	FindByUserID(userID uint) ([]*model.ThirdPartyUserMeta, error)

	// Update updates a third-party user association
	Update(thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error)

	// Delete deletes a third-party user association
	Delete(id uint) error

	// DeleteByUserID deletes by user ID
	DeleteByUserID(userID uint) error
}

func RegisterThirdPartyUserDao(impl ThirdPartyUserDao) {
	di.Provide[ThirdPartyUserDao]("thirdPartyUser.dao", impl)
}

func GetThirdPartyUserDao() ThirdPartyUserDao {
	return di.Require[ThirdPartyUserDao]("thirdPartyUser.dao")
}
