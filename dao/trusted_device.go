package dao

import (
	"context"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// TrustedDeviceDao defines the trusted device data access interface.
type TrustedDeviceDao interface {
	// Create creates a trusted device
	Create(ctx context.Context, trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error)

	// FindByID finds by ID
	FindByID(ctx context.Context, id uint) (*model.TrustedDevice, error)

	// FindByUserIDAndFingerprint finds by user ID and device fingerprint
	FindByUserIDAndFingerprint(ctx context.Context, userID uint, deviceFingerprint string) (*model.TrustedDevice, error)

	// FindByUserID finds by user ID
	FindByUserID(ctx context.Context, userID uint) ([]*model.TrustedDevice, error)

	// Update updates a trusted device
	Update(ctx context.Context, trustedDevice *model.TrustedDevice) (*model.TrustedDevice, error)

	// Delete deletes a trusted device
	Delete(ctx context.Context, id uint) error

	// DeleteByUserID deletes by user ID
	DeleteByUserID(ctx context.Context, userID uint) error
}

func RegisterTrustedDeviceDao(impl TrustedDeviceDao) {
	di.Provide[TrustedDeviceDao]("trustedDevice.dao", impl)
}

func GetTrustedDeviceDao() TrustedDeviceDao {
	return di.Require[TrustedDeviceDao]("trustedDevice.dao")
}
