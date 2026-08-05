package third_party_user

import (
	"context"

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

func (r *thirdPartyUserImpl) Create(ctx context.Context, thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error) {
	err := framedb.GetDBFrom(ctx).Create(thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) FindByID(ctx context.Context, id uint) (*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMeta model.ThirdPartyUserMeta
	err := framedb.GetDBFrom(ctx).First(&thirdPartyUserMeta, id).Error
	if err != nil {
		return nil, err
	}
	return &thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) FindByOpenIDAndProvider(ctx context.Context, openID string, provider string) (*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMeta model.ThirdPartyUserMeta
	err := framedb.GetDBFrom(ctx).Where("open_id = ? AND provider = ?", openID, provider).First(&thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return &thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) FindByUnionID(ctx context.Context, unionID string) (*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMeta model.ThirdPartyUserMeta
	err := framedb.GetDBFrom(ctx).Where("union_id = ?", unionID).First(&thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return &thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) FindByUserID(ctx context.Context, userID uint) ([]*model.ThirdPartyUserMeta, error) {
	var thirdPartyUserMetas []*model.ThirdPartyUserMeta
	err := framedb.GetDBFrom(ctx).Where("user_id = ?", userID).Find(&thirdPartyUserMetas).Error
	if err != nil {
		return nil, err
	}
	return thirdPartyUserMetas, nil
}

func (r *thirdPartyUserImpl) Update(ctx context.Context, thirdPartyUserMeta *model.ThirdPartyUserMeta) (*model.ThirdPartyUserMeta, error) {
	err := framedb.GetDBFrom(ctx).Save(thirdPartyUserMeta).Error
	if err != nil {
		return nil, err
	}
	return thirdPartyUserMeta, nil
}

func (r *thirdPartyUserImpl) Delete(ctx context.Context, id uint) error {
	return framedb.GetDBFrom(ctx).Delete(&model.ThirdPartyUserMeta{}, id).Error
}

func (r *thirdPartyUserImpl) DeleteByUserID(ctx context.Context, userID uint) error {
	return framedb.GetDBFrom(ctx).Where("user_id = ?", userID).Delete(&model.ThirdPartyUserMeta{}).Error
}
