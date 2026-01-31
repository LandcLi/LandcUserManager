package dao

import (
	"github.com/LandcLi/LandcUserManager/model"
)

var (
	thirdPartyUserDaoImpl thirdPartyUserDao
)

func RegisterThirdPartyUserDaoImpl(impl thirdPartyUserDao) {
	thirdPartyUserDaoImpl = impl
}

func GetThirdPartyUserDao() thirdPartyUserDao {
	return thirdPartyUserDaoImpl
}

// thirdPartyUserDao 第三方用户仓库接口
type thirdPartyUserDao interface {
	// Create 创建第三方用户关联
	Create(thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error)

	// FindByID 根据ID查找第三方用户关联
	FindByID(id uint) (*model.ThirdPartyUserMeta, error)

	// FindByOpenIDAndProvider 根据OpenID和提供商查找第三方用户关联
	FindByOpenIDAndProvider(openID string, provider string) (*model.ThirdPartyUserMeta, error)

	// FindByUnionID 根据UnionID查找第三方用户关联
	FindByUnionID(unionID string) (*model.ThirdPartyUserMeta, error)

	// FindByUserID 根据用户ID查找第三方用户关联列表
	FindByUserID(userID uint) ([]*model.ThirdPartyUserMeta, error)

	// Update 更新第三方用户关联
	Update(thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error)

	// Delete 删除第三方用户关联
	Delete(id uint) error

	// DeleteByUserID 根据用户ID删除第三方用户关联
	DeleteByUserID(userID uint) error
}
