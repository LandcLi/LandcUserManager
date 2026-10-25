package login_history

import (
	"context"
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

func (r *loginHistoryImpl) Create(ctx context.Context, loginHistory *model.LoginHistory) (*model.LoginHistory, error) {
	err := framedb.GetDBFrom(ctx).Create(loginHistory).Error
	if err != nil {
		return nil, err
	}
	return loginHistory, nil
}

func (r *loginHistoryImpl) FindByID(ctx context.Context, id uint) (*model.LoginHistory, error) {
	var loginHistory model.LoginHistory
	err := framedb.GetDBFrom(ctx).First(&loginHistory, id).Error
	if err != nil {
		return nil, err
	}
	return &loginHistory, nil
}

func (r *loginHistoryImpl) FindByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.LoginHistory, int64, error) {
	var loginHistories []*model.LoginHistory
	var total int64

	err := framedb.GetDBFrom(ctx).Model(&model.LoginHistory{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = framedb.GetDBFrom(ctx).Where("user_id = ?", userID).Order("login_time DESC").Offset(offset).Limit(pageSize).Find(&loginHistories).Error
	if err != nil {
		return nil, 0, err
	}

	return loginHistories, total, nil
}

// FindLatestByUserID 返回该用户最近一条登录历史（按登录时间倒序取第一条）。
func (r *loginHistoryImpl) FindLatestByUserID(ctx context.Context, userID uint) (*model.LoginHistory, error) {
	var loginHistory model.LoginHistory
	err := framedb.GetDBFrom(ctx).Where("user_id = ?", userID).Order("login_time DESC").First(&loginHistory).Error
	if err != nil {
		return nil, err
	}
	return &loginHistory, nil
}

func (r *loginHistoryImpl) Delete(ctx context.Context, id uint) error {
	return framedb.GetDBFrom(ctx).Delete(&model.LoginHistory{}, id).Error
}

func (r *loginHistoryImpl) DeleteByUserID(ctx context.Context, userID uint) error {
	return framedb.GetDBFrom(ctx).Where("user_id = ?", userID).Delete(&model.LoginHistory{}).Error
}

func (r *loginHistoryImpl) DeleteOldRecords(ctx context.Context, days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	return framedb.GetDBFrom(ctx).Where("login_time < ?", cutoffTime).Delete(&model.LoginHistory{}).Error
}
