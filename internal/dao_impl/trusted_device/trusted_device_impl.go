package trusted_device

import (
	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/utility/db"

	"gorm.io/gorm"
)

func init() {
	if dao.GetTrustedDevice() == nil {
		dao.SetTrustedDeviceImpl(NewTrustedDeviceImpl())
	}
}

// trustedDeviceImpl 信任设备仓库实现
type trustedDeviceImpl struct {
	orm *gorm.DB
}

// NewTrustedDeviceImpl 创建信任设备仓库实例
func NewTrustedDeviceImpl() dao.TrustedDevice {
	return &trustedDeviceImpl{
		orm: db.GetDB(),
	}
}

// Create 创建信任设备
func (r *trustedDeviceImpl) Create(trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error) {
	err := r.orm.Create(trustedDevice).Error
	if err != nil {
		return nil, err
	}
	return trustedDevice, nil
}

// FindByID 根据ID查找信任设备
func (r *trustedDeviceImpl) FindByID(id uint) (*model.TrustedDevice, error) {
	var trustedDevice model.TrustedDevice
	err := r.orm.First(&trustedDevice, id).Error
	if err != nil {
		return nil, err
	}
	return &trustedDevice, nil
}

// FindByUserIDAndFingerprint 根据用户ID和设备指纹查找信任设备
func (r *trustedDeviceImpl) FindByUserIDAndFingerprint(userID uint, deviceFingerprint string) (*model.TrustedDevice, error) {
	var trustedDevice model.TrustedDevice
	err := r.orm.Where("user_id = ? AND device_fingerprint = ?", userID, deviceFingerprint).First(&trustedDevice).Error
	if err != nil {
		return nil, err
	}
	return &trustedDevice, nil
}

// FindByUserID 根据用户ID查找信任设备列表
func (r *trustedDeviceImpl) FindByUserID(userID uint) ([]*model.TrustedDevice, error) {
	var trustedDevices []*model.TrustedDevice
	err := r.orm.Where("user_id = ?", userID).Find(&trustedDevices).Error
	if err != nil {
		return nil, err
	}
	return trustedDevices, nil
}

// Update 更新信任设备
func (r *trustedDeviceImpl) Update(trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error) {
	err := r.orm.Save(trustedDevice).Error
	if err != nil {
		return nil, err
	}
	return trustedDevice, nil
}

// Delete 删除信任设备
func (r *trustedDeviceImpl) Delete(id uint) error {
	return r.orm.Delete(&model.TrustedDevice{}, id).Error
}

// DeleteByUserID 根据用户ID删除信任设备
func (r *trustedDeviceImpl) DeleteByUserID(userID uint) error {
	return r.orm.Where("user_id = ?", userID).Delete(&model.TrustedDevice{}).Error
}
