package third_party_user

import (
	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/utility/db"
	"github.com/LandcLi/landc-go/frame/pkg/di"

	"gorm.io/gorm"
)

func init() {
	di.Provide[dao.ThirdPartyUserDao]("thirdPartyUser.dao", NewThirdPartyUserImpl())
}

// thirdPartyUserImpl 第三方用户仓库实现
type thirdPartyUserImpl struct {
	orm *gorm.DB
}

// NewThirdPartyUserImpl 创建第三方用户仓库实例
func NewThirdPartyUserImpl() *thirdPartyUserImpl {
	return &thirdPartyUserImpl{
		orm: db.GetDB(),
	}
}

// Create 创建第三方用户关联
func (r *thirdPartyUserImpl) Create(thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error) {
	err := r.orm.Create(thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return thirdPartyUserMeta, nil
}

// FindByID 根据ID查找第三方用户关联
func (r *thirdPartyUserImpl) FindByID(id uint) (*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMeta model.ThirdPartyUserMeta
	err := r.orm.First(&thirdPartyUserMeta, id).Error
	if err != nil {
		return nil, err
	}
	return &thirdPartyUserMeta, nil
}

// FindByOpenIDAndProvider 根据OpenID和提供商查找第三方用户关联
func (r *thirdPartyUserImpl) FindByOpenIDAndProvider(openID string, provider string) (*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMeta model.ThirdPartyUserMeta
	err := r.orm.Where("open_id = ? AND provider = ?", openID, provider).First(&thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return &thirdPartyUserMeta, nil
}

// FindByUnionID 根据UnionID查找第三方用户关联
func (r *thirdPartyUserImpl) FindByUnionID(unionID string) (*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMeta model.ThirdPartyUserMeta
	err := r.orm.Where("union_id = ?", unionID).First(&thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return &thirdPartyUserMeta, nil
}

// FindByUserID 根据用户ID查找第三方用户关联列表
func (r *thirdPartyUserImpl) FindByUserID(userID uint) ([]*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMetas []*model.ThirdPartyUserMeta
	err := r.orm.Where("user_id = ?", userID).Find(&thirdPartyUserMetas).Error
	if err != nil {
		return nil, err
	}
	return thirdPartyUserMetas, nil
}

// Update 更新第三方用户关联
func (r *thirdPartyUserImpl) Update(thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error) {
	err := r.orm.Save(thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return thirdPartyUserMeta, nil
}

// Delete 删除第三方用户关联
func (r *thirdPartyUserImpl) Delete(id uint) error {
	return r.orm.Delete(&model.ThirdPartyUserMeta{}, id).Error
}

// DeleteByUserID 根据用户ID删除第三方用户关联
func (r *thirdPartyUserImpl) DeleteByUserID(userID uint) error {
	return r.orm.Where("user_id = ?", userID).Delete(&model.ThirdPartyUserMeta{}).Error
}
