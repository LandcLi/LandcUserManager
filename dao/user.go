package dao

import (
	"context"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// UserDao defines the user data access interface.
type UserDao interface {
	// Create creates a user
	Create(ctx context.Context, user *model.User) (*model.User, error)

	// FindByID finds user by ID
	FindByID(ctx context.Context, id uint) (*model.User, error)

	// FindByUsername finds user by username
	FindByUsername(ctx context.Context, username string) (*model.User, error)

	// FindByUsernameFromDB finds user by username directly from DB, skipping cache
	FindByUsernameFromDB(ctx context.Context, username string) (*model.User, error)

	// FindByEmailFromDB finds user by email directly from DB, skipping cache
	FindByEmailFromDB(ctx context.Context, email string) (*model.User, error)

	// FindByPhoneFromDB finds user by phone directly from DB, skipping cache
	FindByPhoneFromDB(ctx context.Context, phone string) (*model.User, error)

	// FindByEmail finds user by email
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// FindByPhone finds user by phone
	FindByPhone(ctx context.Context, phone string) (*model.User, error)

	// Update updates a user
	Update(ctx context.Context, user *model.User) (*model.User, error)

	// Delete deletes a user
	Delete(ctx context.Context, id uint) error

	// ClearUserCache clears user cache
	ClearUserCache(ctx context.Context, user *model.User) error

	// List gets user list
	List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error)
}

func RegisterUserDao(impl UserDao) {
	di.Provide[UserDao]("user.dao", impl)
}

func GetUserDao() UserDao {
	return di.Require[UserDao]("user.dao")
}
