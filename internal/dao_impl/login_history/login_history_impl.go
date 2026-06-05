package login_history

import (
	"time"

	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/frame/pkg/di"

)

func init() {
	di.Provide[dao.LoginHistoryDao]("loginHistory.dao", NewLoginHistoryImpl())
}

type loginHistoryImpl struct {
}

func NewLoginHistoryImpl() *loginHistoryImpl {
	return &loginHistoryImpl{}
}


func (r *loginHistoryImpl) Create(loginHistory *model.LoginHistory) (*model.LoginHistory, error) {
	err := framedb.GetDB().Create(loginHistory).Error
	if err != nil {
		return nil, err
	}
	return loginHistory, nil
}

func (r *loginHistoryImpl) FindByID(id uint) (*model.LoginHistory, error) {
	var loginHistory model.LoginHistory
	err := framedb.GetDB().First(&loginHistory, id).Error
	if err != nil {
		return nil, err
	}
	return &loginHistory, nil
}

func (r *loginHistoryImpl) FindByUserID(userID uint, page, pageSize int) ([]*model.LoginHistory, int64, error) {
	var loginHistories []*model.LoginHistory
	var total int64

	err := framedb.GetDB().Model(&model.LoginHistory{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = framedb.GetDB().Where("user_id = ?", userID).Order("login_time DESC").Offset(offset).Limit(pageSize).Find(&loginHistories).Error
	if err != nil {
		return nil, 0, err
	}

	return loginHistories, total, nil
}

func (r *loginHistoryImpl) Delete(id uint) error {
	return framedb.GetDB().Delete(&model.LoginHistory{}, id).Error
}

func (r *loginHistoryImpl) DeleteByUserID(userID uint) error {
	return framedb.GetDB().Where("user_id = ?", userID).Delete(&model.LoginHistory{}).Error
}

func (r *loginHistoryImpl) DeleteOldRecords(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	return framedb.GetDB().Where("login_time < ?", cutoffTime).Delete(&model.LoginHistory{}).Error
}
