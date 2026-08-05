package user

import (
	"context"
	"fmt"
	"time"

	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/frame/pkg/di"
)

func init() {
	di.Provide[dao.UserDao]("user.dao", NewUserImpl())
}

type userImpl struct{}

func NewUserImpl() *userImpl {
	return &userImpl{}
}

func (r *userImpl) cache(ctx context.Context) framecache.Cache {
	return framecache.GetCacheFrom(ctx)
}

// Create 创建用户
func (r *userImpl) Create(ctx context.Context, user *model.User) (*model.User, error) {
	err := framedb.GetDBFrom(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID 根据ID查找用户（带缓存）
func (r *userImpl) FindByID(ctx context.Context, id uint) (*model.User, error) {
	c := r.cache(ctx)
	if c != nil {
		cacheKey := fmt.Sprintf("user:id:%d", id)
		var user model.User
		if err := c.GetObject(ctx, cacheKey, &user); err == nil {
			return &user, nil
		}
	}

	var user model.User
	err := framedb.GetDBFrom(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}

	if c := r.cache(ctx); c != nil {
		cacheKey := fmt.Sprintf("user:id:%d", id)
		_ = c.SetObject(ctx, cacheKey, user, time.Hour)
	}

	return &user, nil
}

// FindByUsername 根据用户名查找用户（带缓存）
func (r *userImpl) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	c := r.cache(ctx)
	if c != nil {
		cacheKey := fmt.Sprintf("user:username:%s", username)
		var user model.User
		if err := c.GetObject(ctx, cacheKey, &user); err == nil {
			return &user, nil
		}
	}

	var user model.User
	err := framedb.GetDBFrom(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	if c := r.cache(ctx); c != nil {
		cacheKey := fmt.Sprintf("user:username:%s", username)
		_ = c.SetObject(ctx, cacheKey, user, time.Hour)
	}

	return &user, nil
}

// FindByUsernameFromDB 跳过缓存，直接查库
func (r *userImpl) FindByUsernameFromDB(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := framedb.GetDBFrom(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmailFromDB 跳过缓存，直接查库
func (r *userImpl) FindByEmailFromDB(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := framedb.GetDBFrom(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByPhoneFromDB 跳过缓存，直接查库
func (r *userImpl) FindByPhoneFromDB(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := framedb.GetDBFrom(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户（带缓存）
func (r *userImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	c := r.cache(ctx)
	if c != nil {
		cacheKey := fmt.Sprintf("user:email:%s", email)
		var user model.User
		if err := c.GetObject(ctx, cacheKey, &user); err == nil {
			return &user, nil
		}
	}

	var user model.User
	err := framedb.GetDBFrom(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	if c := r.cache(ctx); c != nil {
		cacheKey := fmt.Sprintf("user:email:%s", email)
		_ = c.SetObject(ctx, cacheKey, user, time.Hour)
	}

	return &user, nil
}

// FindByPhone 根据手机号查找用户（带缓存）
func (r *userImpl) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	c := r.cache(ctx)
	if c != nil {
		cacheKey := fmt.Sprintf("user:phone:%s", phone)
		var user model.User
		if err := c.GetObject(ctx, cacheKey, &user); err == nil {
			return &user, nil
		}
	}

	var user model.User
	err := framedb.GetDBFrom(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}

	if c := r.cache(ctx); c != nil {
		cacheKey := fmt.Sprintf("user:phone:%s", phone)
		_ = c.SetObject(ctx, cacheKey, user, time.Hour)
	}

	return &user, nil
}

// Update 更新用户并刷新缓存
func (r *userImpl) Update(ctx context.Context, user *model.User) (*model.User, error) {
	err := framedb.GetDBFrom(ctx).Save(user).Error
	if err != nil {
		return nil, err
	}

	c := r.cache(ctx)
	if c != nil {
		if user.ID != 0 {
			_ = c.SetObject(ctx, fmt.Sprintf("user:id:%d", user.ID), user, time.Hour)
		}
		if user.Username != "" {
			_ = c.SetObject(ctx, fmt.Sprintf("user:username:%s", user.Username), user, time.Hour)
		}
		if user.Email != "" {
			_ = c.SetObject(ctx, fmt.Sprintf("user:email:%s", user.Email), user, time.Hour)
		}
		if user.Phone != "" {
			_ = c.SetObject(ctx, fmt.Sprintf("user:phone:%s", user.Phone), user, time.Hour)
		}
	}

	return user, nil
}

// Delete 删除用户并清除缓存
func (r *userImpl) Delete(ctx context.Context, id uint) error {
	var user model.User
	if err := framedb.GetDBFrom(ctx).First(&user, id).Error; err == nil {
		if c := r.cache(ctx); c != nil {
			_ = c.Delete(ctx, fmt.Sprintf("user:id:%d", id),
				fmt.Sprintf("user:username:%s", user.Username),
				fmt.Sprintf("user:email:%s", user.Email),
				fmt.Sprintf("user:phone:%s", user.Phone),
			)
		}
	}

	return framedb.GetDBFrom(ctx).Delete(&model.User{}, id).Error
}

// ClearUserCache 清除用户缓存
func (r *userImpl) ClearUserCache(ctx context.Context, user *model.User) error {
	c := r.cache(ctx)
	if c == nil {
		return nil
	}

	keys := []string{fmt.Sprintf("user:id:%d", user.ID)}
	if user.Username != "" {
		keys = append(keys, fmt.Sprintf("user:username:%s", user.Username))
	}
	if user.Email != "" {
		keys = append(keys, fmt.Sprintf("user:email:%s", user.Email))
	}
	if user.Phone != "" {
		keys = append(keys, fmt.Sprintf("user:phone:%s", user.Phone))
	}

	return c.Delete(ctx, keys...)
}

// List 分页获取用户列表
func (r *userImpl) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	err := framedb.GetDBFrom(ctx).Model(&model.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = framedb.GetDBFrom(ctx).Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
