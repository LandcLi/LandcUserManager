package dao

import (
	"github.com/LandcLi/LandcUserManager/model"
)

var (
	loginHistoryDaoImpl loginHistoryDao
)

func RegisterLoginHistoryDaoImpl(impl loginHistoryDao) {
	loginHistoryDaoImpl = impl
}

func GetLoginHistoryDao() loginHistoryDao {
	return loginHistoryDaoImpl
}

// loginHistoryDao 登录历史仓库接口
type loginHistoryDao interface {
	// Create 创建登录历史
	Create(loginHistory *model.LoginHistory) (*model.LoginHistory, error)

	// FindByID 根据ID查找登录历史
	FindByID(id uint) (*model.LoginHistory, error)

	// FindByUserID 根据用户ID查找登录历史
	FindByUserID(userID uint, page, pageSize int) ([]*model.LoginHistory, int64, error)

	// Delete 删除登录历史
	Delete(id uint) error

	// DeleteByUserID 根据用户ID删除登录历史
	DeleteByUserID(userID uint) error

	// DeleteOldRecords 删除旧记录
	DeleteOldRecords(days int) error
}
