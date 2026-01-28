package dao

import (
	"github.com/LandcLi/LandcUserManager/model"
)

var (
	userDao User
)

func SetUserImpl(impl User) {
	userDao = impl
}

func GetUser() User {
	return userDao
}

// User 用户仓库接口
type User interface {
	// Create 创建用户
	Create(user *model.User) (*model.User, error)

	// FindByID 根据ID查找用户
	FindByID(id uint) (*model.User, error)

	// FindByUsername 根据用户名查找用户
	FindByUsername(username string) (*model.User, error)

	// FindByUsernameFromDB 根据用户名从数据库直接查找用户，跳过缓存
	FindByUsernameFromDB(username string) (*model.User, error)

	// FindByEmailFromDB 根据邮箱从数据库直接查找用户，跳过缓存
	FindByEmailFromDB(email string) (*model.User, error)

	// FindByPhoneFromDB 根据手机号从数据库直接查找用户，跳过缓存
	FindByPhoneFromDB(phone string) (*model.User, error)

	// FindByEmail 根据邮箱查找用户
	FindByEmail(email string) (*model.User, error)

	// FindByPhone 根据手机号查找用户
	FindByPhone(phone string) (*model.User, error)

	// Update 更新用户
	Update(user *model.User) (*model.User, error)

	// Delete 删除用户
	Delete(id uint) error

	// ClearUserCache 清除用户缓存
	ClearUserCache(user *model.User) error

	// List 获取用户列表
	List(page, pageSize int) ([]*model.User, int64, error)
}
