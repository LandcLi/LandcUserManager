package login_history

import (
	"time"

	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/utility/db"
	"github.com/LandcLi/landc-go/mvc/pkg/di"

	"gorm.io/gorm"
)

func init() {
	di.Provide[dao.LoginHistoryDao]("loginHistory.dao", NewLoginHistoryImpl())
}

// loginHistoryImpl 登录历史仓库实现
type loginHistoryImpl struct {
	orm *gorm.DB
}

// NewLoginHistoryImpl 创建登录历史仓库实例
func NewLoginHistoryImpl() *loginHistoryImpl {
	return &loginHistoryImpl{
		orm: db.GetDB(),
	}
}

// Create 创建登录历史
func (r *loginHistoryImpl) Create(loginHistory *model.LoginHistory) (*model.LoginHistory, error) {
	err := r.orm.Create(loginHistory).Error
	if err != nil {
		return nil, err
	}
	return loginHistory, nil
}

// FindByID 根据ID查找登录历史
func (r *loginHistoryImpl) FindByID(id uint) (*model.LoginHistory, error) {
	var loginHistory model.LoginHistory
	err := r.orm.First(&loginHistory, id).Error
	if err != nil {
		return nil, err
	}
	return &loginHistory, nil
}

// FindByUserID 根据用户ID查找登录历史
func (r *loginHistoryImpl) FindByUserID(userID uint, page, pageSize int) ([]*model.LoginHistory, int64, error) {
	var loginHistories []*model.LoginHistory
	var total int64

	// 计算总数
	err := r.orm.Model(&model.LoginHistory{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = r.orm.Where("user_id = ?", userID).Order("login_time DESC").Offset(offset).Limit(pageSize).Find(&loginHistories).Error
	if err != nil {
		return nil, 0, err
	}

	return loginHistories, total, nil
}

// Delete 删除登录历史
func (r *loginHistoryImpl) Delete(id uint) error {
	return r.orm.Delete(&model.LoginHistory{}, id).Error
}

// DeleteByUserID 根据用户ID删除登录历史
func (r *loginHistoryImpl) DeleteByUserID(userID uint) error {
	return r.orm.Where("user_id = ?", userID).Delete(&model.LoginHistory{}).Error
}

// DeleteOldRecords 删除旧记录
func (r *loginHistoryImpl) DeleteOldRecords(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	return r.orm.Where("login_time < ?", cutoffTime).Delete(&model.LoginHistory{}).Error
}
