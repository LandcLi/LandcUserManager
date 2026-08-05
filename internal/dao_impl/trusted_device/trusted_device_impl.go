package trusted_device

import (
	"context"

	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

func init() {
	di.Provide[dao.TrustedDeviceDao]("trustedDevice.dao", NewTrustedDeviceImpl())
}

type trustedDeviceImpl struct {
}

func NewTrustedDeviceImpl() *trustedDeviceImpl {
	return &trustedDeviceImpl{}
}

func (r *trustedDeviceImpl) Create(ctx context.Context, trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error) {
	err := framedb.GetDBFrom(ctx).Create(trustedDevice).Error
	if err != nil {
		return nil, err
	}
	return trustedDevice, nil
}

func (r *trustedDeviceImpl) FindByID(ctx context.Context, id uint) (*model.TrustedDevice, error) {
	var trustedDevice model.TrustedDevice
	err := framedb.GetDBFrom(ctx).First(&trustedDevice, id).Error
	if err != nil {
		return nil, err
	}
	return &trustedDevice, nil
}

func (r *trustedDeviceImpl) FindByUserIDAndFingerprint(ctx context.Context, userID uint, deviceFingerprint string) (*model.TrustedDevice, error) {
	var trustedDevice model.TrustedDevice
	err := framedb.GetDBFrom(ctx).Where("user_id = ? AND device_fingerprint = ?", userID, deviceFingerprint).First(&trustedDevice).Error
	if err != nil {
		return nil, err
	}
	return &trustedDevice, nil
}

func (r *trustedDeviceImpl) FindByUserID(ctx context.Context, userID uint) ([]*model.TrustedDevice, error) {
	var trustedDevices []*model.TrustedDevice
	err := framedb.GetDBFrom(ctx).Where("user_id = ?", userID).Find(&trustedDevices).Error
	if err != nil {
		return nil, err
	}
	return trustedDevices, nil
}

func (r *trustedDeviceImpl) Update(ctx context.Context, trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error) {
	err := framedb.GetDBFrom(ctx).Save(trustedDevice).Error
	if err != nil {
		return nil, err
	}
	return trustedDevice, nil
}

func (r *trustedDeviceImpl) Delete(ctx context.Context, id uint) error {
	return framedb.GetDBFrom(ctx).Delete(&model.TrustedDevice{}, id).Error
}

func (r *trustedDeviceImpl) DeleteByUserID(ctx context.Context, userID uint) error {
	return framedb.GetDBFrom(ctx).Where("user_id = ?", userID).Delete(&model.TrustedDevice{}).Error
}
