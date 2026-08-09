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
			// 缓存中的 Password 因 json:"-" 被序列化排除；
			// 敏感操作（修改密码等）需要完整记录，缺失时回源 DB 补齐。
			if user.Password == "" {
				if dbErr := framedb.GetDBFrom(ctx).First(&user, id).Error; dbErr != nil {
					return nil, dbErr
				}
			}
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

// Update 更新用户并刷新缓存。
// 使用 Updates（仅非零字段）而非 Save（全字段覆盖），
// 避免因缓存/业务传入的零值字段（如 Password 为空）覆盖数据库中的既有值。
func (r *userImpl) Update(ctx context.Context, user *model.User) (*model.User, error) {
	if user.ID == 0 {
		return nil, fmt.Errorf("user update requires non-zero ID")
	}
	err := framedb.GetDBFrom(ctx).Updates(user).Error
	if err != nil {
		return nil, err
	}

	// 重新读取完整记录（含密码），确保返回与缓存一致
	var fresh model.User
	if err := framedb.GetDBFrom(ctx).First(&fresh, user.ID).Error; err != nil {
		return nil, err
	}

	c := r.cache(ctx)
	if c != nil {
		if fresh.ID != 0 {
			_ = c.SetObject(ctx, fmt.Sprintf("user:id:%d", fresh.ID), fresh, time.Hour)
		}
		if fresh.Username != "" {
			_ = c.SetObject(ctx, fmt.Sprintf("user:username:%s", fresh.Username), fresh, time.Hour)
		}
		if fresh.Email != "" {
			_ = c.SetObject(ctx, fmt.Sprintf("user:email:%s", fresh.Email), fresh, time.Hour)
		}
		if fresh.Phone != "" {
			_ = c.SetObject(ctx, fmt.Sprintf("user:phone:%s", fresh.Phone), fresh, time.Hour)
		}
	}

	return &fresh, nil
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
