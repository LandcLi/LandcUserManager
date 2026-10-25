package dao

import (
	"context"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

// LoginHistoryDao defines the login history data access interface.
type LoginHistoryDao interface {
	// Create creates a login history record
	Create(ctx context.Context, loginHistory *model.LoginHistory) (*model.LoginHistory, error)

	// FindByID finds by ID
	FindByID(ctx context.Context, id uint) (*model.LoginHistory, error)

	// FindByUserID finds by user ID with pagination
	FindByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.LoginHistory, int64, error)

	// FindLatestByUserID 返回该用户最近一条登录历史，无记录时返回 error。
	// 用于异地登录判定（对比最近一次登录 IP）。
	FindLatestByUserID(ctx context.Context, userID uint) (*model.LoginHistory, error)

	// Delete deletes a login history record
	Delete(ctx context.Context, id uint) error

	// DeleteByUserID deletes by user ID
	DeleteByUserID(ctx context.Context, userID uint) error

	// DeleteOldRecords deletes records older than specified days
	DeleteOldRecords(ctx context.Context, days int) error
}

func RegisterLoginHistoryDao(impl LoginHistoryDao) {
	di.Provide[LoginHistoryDao]("loginHistory.dao", impl)
}

func GetLoginHistoryDao() LoginHistoryDao {
	return di.Require[LoginHistoryDao]("loginHistory.dao")
}
