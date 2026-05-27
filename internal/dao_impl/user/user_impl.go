package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/utility/db"
	"github.com/LandcLi/landc-go/mvc/pkg/di"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func init() {
	di.Provide[dao.UserDao]("user.dao", NewUserImpl())
}

// userImpl 用户仓库实现
type userImpl struct {
	orm         *gorm.DB
	redisClient redis.UniversalClient
}

// NewUserImpl 创建用户仓库实例
func NewUserImpl() *userImpl {
	return &userImpl{
		orm:         db.GetDB(),
		redisClient: db.GetRedis(),
	}
}

// Create 创建用户
func (r *userImpl) Create(user *model.User) (*model.User, error) {
	err := r.orm.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID 根据ID查找用户
func (r *userImpl) FindByID(id uint) (*model.User, error) {
	// 先从缓存中查找
	if r.redisClient != nil {
		cacheKey := fmt.Sprintf("user:id:%d", id)
		data, err := r.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var user model.User
			if err := json.Unmarshal(data, &user); err == nil {
				return &user, nil
			}
		}
	}

	// 缓存未命中，从数据库中查找
	var user model.User
	err := r.orm.First(&user, id).Error
	if err != nil {
		return nil, err
	}

	// 存入缓存
	if r.redisClient != nil {
		cacheKey := fmt.Sprintf("user:id:%d", id)
		data, err := json.Marshal(user)
		if err == nil {
			// 设置缓存，过期时间为1小时
			r.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
		}
	}

	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *userImpl) FindByUsername(username string) (*model.User, error) {
	// 先从缓存中查找
	if r.redisClient != nil {
		cacheKey := fmt.Sprintf("user:username:%s", username)
		data, err := r.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var user model.User
			if err := json.Unmarshal(data, &user); err == nil {
				return &user, nil
			}
		}
	}

	// 缓存未命中，从数据库中查找
	var user model.User
	err := r.orm.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	// 存入缓存
	if r.redisClient != nil {
		cacheKey := fmt.Sprintf("user:username:%s", username)
		data, err := json.Marshal(user)
		if err == nil {
			// 设置缓存，过期时间为1小时
			r.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
		}
	}

	return &user, nil
}

// FindByUsernameFromDB 根据用户名从数据库直接查找用户，跳过缓存
func (r *userImpl) FindByUsernameFromDB(username string) (*model.User, error) {
	var user model.User
	err := r.orm.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmailFromDB 根据邮箱从数据库直接查找用户，跳过缓存
func (r *userImpl) FindByEmailFromDB(email string) (*model.User, error) {
	var user model.User
	err := r.orm.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByPhoneFromDB 根据手机号从数据库直接查找用户，跳过缓存
func (r *userImpl) FindByPhoneFromDB(phone string) (*model.User, error) {
	var user model.User
	err := r.orm.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *userImpl) FindByEmail(email string) (*model.User, error) {
	// 先从缓存中查找
	if r.redisClient != nil {
		cacheKey := fmt.Sprintf("user:email:%s", email)
		data, err := r.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var user model.User
			if err := json.Unmarshal(data, &user); err == nil {
				return &user, nil
			}
		}
	}

	// 缓存未命中，从数据库中查找
	var user model.User
	err := r.orm.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	// 存入缓存
	if r.redisClient != nil {
		cacheKey := fmt.Sprintf("user:email:%s", email)
		data, err := json.Marshal(user)
		if err == nil {
			// 设置缓存，过期时间为1小时
			r.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
		}
	}

	return &user, nil
}

// FindByPhone 根据手机号查找用户
func (r *userImpl) FindByPhone(phone string) (*model.User, error) {
	// 先从缓存中查找
	if r.redisClient != nil {
		cacheKey := fmt.Sprintf("user:phone:%s", phone)
		data, err := r.redisClient.Get(context.Background(), cacheKey).Bytes()
		if err == nil {
			var user model.User
			if err := json.Unmarshal(data, &user); err == nil {
				return &user, nil
			}
		}
	}

	// 缓存未命中，从数据库中查找
	var user model.User
	err := r.orm.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}

	// 存入缓存
	if r.redisClient != nil {
		cacheKey := fmt.Sprintf("user:phone:%s", phone)
		data, err := json.Marshal(user)
		if err == nil {
			// 设置缓存，过期时间为1小时
			r.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
		}
	}

	return &user, nil
}

// Update 更新用户
func (r *userImpl) Update(user *model.User) (*model.User, error) {
	err := r.orm.Save(user).Error
	if err != nil {
		return nil, err
	}

	// 更新缓存
	if r.redisClient != nil {
		// 更新ID缓存
		cacheKey := fmt.Sprintf("user:id:%d", user.ID)
		data, err := json.Marshal(user)
		if err == nil {
			r.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
		}

		// 更新用户名缓存
		if user.Username != "" {
			cacheKey := fmt.Sprintf("user:username:%s", user.Username)
			data, err := json.Marshal(user)
			if err == nil {
				r.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
			}
		}

		// 更新邮箱缓存
		if user.Email != "" {
			cacheKey := fmt.Sprintf("user:email:%s", user.Email)
			data, err := json.Marshal(user)
			if err == nil {
				r.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
			}
		}

		// 更新手机号缓存
		if user.Phone != "" {
			cacheKey := fmt.Sprintf("user:phone:%s", user.Phone)
			data, err := json.Marshal(user)
			if err == nil {
				r.redisClient.Set(context.Background(), cacheKey, data, time.Hour)
			}
		}
	}

	return user, nil
}

// Delete 删除用户
func (r *userImpl) Delete(id uint) error {
	// 先获取用户信息，以便删除缓存
	var user model.User
	if err := r.orm.First(&user, id).Error; err == nil {
		// 删除缓存
		if r.redisClient != nil {
			// 删除ID缓存
			cacheKey := fmt.Sprintf("user:id:%d", id)
			r.redisClient.Del(context.Background(), cacheKey)

			// 删除用户名缓存
			if user.Username != "" {
				cacheKey := fmt.Sprintf("user:username:%s", user.Username)
				r.redisClient.Del(context.Background(), cacheKey)
			}

			// 删除邮箱缓存
			if user.Email != "" {
				cacheKey := fmt.Sprintf("user:email:%s", user.Email)
				r.redisClient.Del(context.Background(), cacheKey)
			}

			// 删除手机号缓存
			if user.Phone != "" {
				cacheKey := fmt.Sprintf("user:phone:%s", user.Phone)
				r.redisClient.Del(context.Background(), cacheKey)
			}
		}
	}

	return r.orm.Delete(&model.User{}, id).Error
}

// ClearUserCache 清除用户缓存
func (r *userImpl) ClearUserCache(user *model.User) error {
	// 删除缓存
	if r.redisClient != nil {
		// 删除ID缓存
		cacheKey := fmt.Sprintf("user:id:%d", user.ID)
		r.redisClient.Del(context.Background(), cacheKey)

		// 删除用户名缓存
		if user.Username != "" {
			cacheKey := fmt.Sprintf("user:username:%s", user.Username)
			r.redisClient.Del(context.Background(), cacheKey)
		}

		// 删除邮箱缓存
		if user.Email != "" {
			cacheKey := fmt.Sprintf("user:email:%s", user.Email)
			r.redisClient.Del(context.Background(), cacheKey)
		}

		// 删除手机号缓存
		if user.Phone != "" {
			cacheKey := fmt.Sprintf("user:phone:%s", user.Phone)
			r.redisClient.Del(context.Background(), cacheKey)
		}
	}

	return nil
}

// List 获取用户列表
func (r *userImpl) List(page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	// 计算总数
	err := r.orm.Model(&model.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = r.orm.Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
