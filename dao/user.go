package dao

import (
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/landc-go/mvc/pkg/di"
)

// UserDao defines the user data access interface.
type UserDao interface {
	// Create creates a user
	Create(user *model.User) (*model.User, error)

	// FindByID finds user by ID
	FindByID(id uint) (*model.User, error)

	// FindByUsername finds user by username
	FindByUsername(username string) (*model.User, error)

	// FindByUsernameFromDB finds user by username directly from DB, skipping cache
	FindByUsernameFromDB(username string) (*model.User, error)

	// FindByEmailFromDB finds user by email directly from DB, skipping cache
	FindByEmailFromDB(email string) (*model.User, error)

	// FindByPhoneFromDB finds user by phone directly from DB, skipping cache
	FindByPhoneFromDB(phone string) (*model.User, error)

	// FindByEmail finds user by email
	FindByEmail(email string) (*model.User, error)

	// FindByPhone finds user by phone
	FindByPhone(phone string) (*model.User, error)

	// Update updates a user
	Update(user *model.User) (*model.User, error)

	// Delete deletes a user
	Delete(id uint) error

	// ClearUserCache clears user cache
	ClearUserCache(user *model.User) error

	// List gets user list
	List(page, pageSize int) ([]*model.User, int64, error)
}

func RegisterUserDao(impl UserDao) {
	di.Provide[UserDao]("user.dao", impl)
}

func GetUserDao() UserDao {
	return di.Require[UserDao]("user.dao")
}
