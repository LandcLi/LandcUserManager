package dao

import (
	"github.com/LandcLi/LandcUserManager/model"
)

var (
	trustedDeviceDaoImpl trustedDeviceDao
)

func RegisterTrustedDeviceDaoImpl(impl trustedDeviceDao) {
	trustedDeviceDaoImpl = impl
}

func GetTrustedDeviceDao() trustedDeviceDao {
	return trustedDeviceDaoImpl
}

// trustedDeviceDao 信任设备仓库接口
type trustedDeviceDao interface {
	// Create 创建信任设备
	Create(trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error)

	// FindByID 根据ID查找信任设备
	FindByID(id uint) (*model.TrustedDevice, error)

	// FindByUserIDAndFingerprint 根据用户ID和设备指纹查找信任设备
	FindByUserIDAndFingerprint(userID uint, deviceFingerprint string) (*model.TrustedDevice, error)

	// FindByUserID 根据用户ID查找信任设备列表
	FindByUserID(userID uint) ([]*model.TrustedDevice, error)

	// Update 更新信任设备
	Update(trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error)

	// Delete 删除信任设备
	Delete(id uint) error

	// DeleteByUserID 根据用户ID删除信任设备
	DeleteByUserID(userID uint) error
}
