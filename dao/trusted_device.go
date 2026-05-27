package dao

import (
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/landc-go/mvc/pkg/di"
)

// TrustedDeviceDao defines the trusted device data access interface.
type TrustedDeviceDao interface {
	// Create creates a trusted device
	Create(trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error)

	// FindByID finds by ID
	FindByID(id uint) (*model.TrustedDevice, error)

	// FindByUserIDAndFingerprint finds by user ID and device fingerprint
	FindByUserIDAndFingerprint(userID uint, deviceFingerprint string) (*model.TrustedDevice, error)

	// FindByUserID finds by user ID
	FindByUserID(userID uint) ([]*model.TrustedDevice, error)

	// Update updates a trusted device
	Update(trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error)

	// Delete deletes a trusted device
	Delete(id uint) error

	// DeleteByUserID deletes by user ID
	DeleteByUserID(userID uint) error
}

func RegisterTrustedDeviceDao(impl TrustedDeviceDao) {
	di.Provide[TrustedDeviceDao]("trustedDevice.dao", impl)
}

func GetTrustedDeviceDao() TrustedDeviceDao {
	return di.Require[TrustedDeviceDao]("trustedDevice.dao")
}
