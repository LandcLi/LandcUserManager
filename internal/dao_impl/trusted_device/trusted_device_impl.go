package trusted_device

import (
	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/frame/pkg/di"

	"gorm.io/gorm"
)

func init() {
	di.Provide[dao.TrustedDeviceDao]("trustedDevice.dao", NewTrustedDeviceImpl())
}

type trustedDeviceImpl struct {
	orm *gorm.DB
}

func NewTrustedDeviceImpl() *trustedDeviceImpl {
	return &trustedDeviceImpl{
		orm: framedb.GetDB(),
	}
}

func (r *trustedDeviceImpl) Create(trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error) {
	err := r.orm.Create(trustedDevice).Error
	if err != nil {
		return nil, err
	}
	return trustedDevice, nil
}

func (r *trustedDeviceImpl) FindByID(id uint) (*model.TrustedDevice, error) {
	var trustedDevice model.TrustedDevice
	err := r.orm.First(&trustedDevice, id).Error
	if err != nil {
		return nil, err
	}
	return &trustedDevice, nil
}

func (r *trustedDeviceImpl) FindByUserIDAndFingerprint(userID uint, deviceFingerprint string) (*model.TrustedDevice, error) {
	var trustedDevice model.TrustedDevice
	err := r.orm.Where("user_id = ? AND device_fingerprint = ?", userID, deviceFingerprint).First(&trustedDevice).Error
	if err != nil {
		return nil, err
	}
	return &trustedDevice, nil
}

func (r *trustedDeviceImpl) FindByUserID(userID uint) ([]*model.TrustedDevice, error) {
	var trustedDevices []*model.TrustedDevice
	err := r.orm.Where("user_id = ?", userID).Find(&trustedDevices).Error
	if err != nil {
		return nil, err
	}
	return trustedDevices, nil
}

func (r *trustedDeviceImpl) Update(trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error) {
	err := r.orm.Save(trustedDevice).Error
	if err != nil {
		return nil, err
	}
	return trustedDevice, nil
}

func (r *trustedDeviceImpl) Delete(id uint) error {
	return r.orm.Delete(&model.TrustedDevice{}, id).Error
}

func (r *trustedDeviceImpl) DeleteByUserID(userID uint) error {
	return r.orm.Where("user_id = ?", userID).Delete(&model.TrustedDevice{}).Error
}
