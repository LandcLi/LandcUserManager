package third_party_user

import (
	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/frame/pkg/di"

)

func init() {
	di.Provide[dao.ThirdPartyUserDao]("thirdPartyUser.dao", NewThirdPartyUserImpl())
}

type thirdPartyUserImpl struct {
}

func NewThirdPartyUserImpl() *thirdPartyUserImpl {
	return &thirdPartyUserImpl{}
}


func (r *thirdPartyUserImpl) Create(thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error) {
	err := framedb.GetDB().Create(thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) FindByID(id uint) (*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMeta model.ThirdPartyUserMeta
	err := framedb.GetDB().First(&thirdPartyUserMeta, id).Error
	if err != nil {
		return nil, err
	}
	return &thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) FindByOpenIDAndProvider(openID string, provider string) (*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMeta model.ThirdPartyUserMeta
	err := framedb.GetDB().Where("open_id = ? AND provider = ?", openID, provider).First(&thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return &thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) FindByUnionID(unionID string) (*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMeta model.ThirdPartyUserMeta
	err := framedb.GetDB().Where("union_id = ?", unionID).First(&thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return &thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) FindByUserID(userID uint) ([]*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMetas []*model.ThirdPartyUserMeta
	err := framedb.GetDB().Where("user_id = ?", userID).Find(&thirdPartyUserMetas).Error
	if err != nil {
		return nil, err
	}
	return thirdPartyUserMetas, nil
}

func (r *thirdPartyUserImpl) Update(thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error) {
	err := framedb.GetDB().Save(thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) Delete(id uint) error {
	return framedb.GetDB().Delete(&model.ThirdPartyUserMeta{}, id).Error
}

func (r *thirdPartyUserImpl) DeleteByUserID(userID uint) error {
	return framedb.GetDB().Where("user_id = ?", userID).Delete(&model.ThirdPartyUserMeta{}).Error
}
