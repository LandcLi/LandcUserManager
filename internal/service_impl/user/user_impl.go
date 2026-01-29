package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LandcLi/LandcGoTools/generate"
	"github.com/LandcLi/LandcGoTools/mail"
	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/service"

	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/utility/client"
	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/LandcLi/LandcUserManager/utility/db"
	"github.com/LandcLi/LandcUserManager/utility/logger"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	if service.GetUserService() == nil {
		service.SetUserService(NewUserServiceImpl())
	}
}

// TODO 扫码登录、生物登录
// UserService 认证服务
type userServiceImpl struct {
	userDao           dao.User
	loginHistoryDao   dao.LoginHistory
	trustedDeviceDao  dao.TrustedDevice
	thirdPartyUserDao dao.ThirdPartyUser
	config            *config.Config
}

// NewUserService 创建认证服务实例
func NewUserServiceImpl() service.UserService {
	// Go 1.20+ 不需要手动初始化随机数种子
	// 系统会自动初始化
	service := &userServiceImpl{
		userDao:           dao.GetUser(),
		loginHistoryDao:   dao.GetLoginHistory(),
		trustedDeviceDao:  dao.GetTrustedDevice(),
		thirdPartyUserDao: dao.GetThirdPartyUser(),
		config:            config.GetConfig(),
	}

	return service
}

// Login 密码验证
func (s *userServiceImpl) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	var (
		user *model.User
		err  error
	)
	switch req.AuthType {
	case "PASSWORD":
		user, err = s.handleAuthByPassword(ctx, req)
	case "VERIFY_CODE":
		user, err = s.handleAuthByVerifyCode(ctx, req)
	case "THIRD_PARTY":
		user, err = s.handleAuthByThirdParty(ctx, req)
	default:
		return nil, errors.New("无效的认证类型")
	}
	if err != nil {
		return nil, err
	}
	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("用户已禁用")
	}

	// 检测信任设备
	isNewDevice := s.detectTrustedDevice(user.ID, req.DeviceFingerprint)

	// 检测异地登录
	isDifferentLocation := s.detectDifferentLocation(user.ID, req.IP)

	// 返回验证结果，包含是否异地登录和是否新设备，是否成功交给SSO系统判断，判断之后会回调LoginRecord进行登陆审计
	return &v1.LoginResponse{
		UserID:            user.ID,
		DifferentLocation: isDifferentLocation,
		IsNewDevice:       isNewDevice,
	}, nil
}
func (s *userServiceImpl) handleAuthByPassword(ctx context.Context, req *v1.LoginRequest) (*model.User, error) {
	var (
		user *model.User
		err  error
	)
	switch req.VerifyType {
	case "EMAIL":
		user, err = s.userDao.FindByEmailFromDB(req.Identifier)
	case "PHONE":
		user, err = s.userDao.FindByPhoneFromDB(req.Identifier)
	case "USERNAME":
		user, err = s.userDao.FindByUsernameFromDB(req.Identifier)
	default:
		return nil, errors.New("无效的验证类型")
	}
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	err = s.checkPassword(user, req.AuthCode)
	if err != nil {
		return nil, errors.New("密码错误")
	}
	return user, nil
}

// 使用验证码登陆时，第一次登陆默认同时注册用户
func (s *userServiceImpl) handleAuthByVerifyCode(ctx context.Context, req *v1.LoginRequest) (*model.User, error) {
	var (
		user *model.User
		err  error
	)
	switch req.VerifyType {
	case "EMAIL":
		user, err = s.userDao.FindByEmailFromDB(req.Identifier)
	case "PHONE":
		user, err = s.userDao.FindByPhoneFromDB(req.Identifier)
	default:
		return nil, errors.New("无效的验证类型")
	}
	if err != nil {
		user, err = s.createUser(req.VerifyType, req.Identifier, "", req.DeviceFingerprint, req.UserAgent, req.IP)
		if err != nil {
			return nil, errors.New("用户不存在")
		}
	}
	err = s.checkVerifyCode(user, req.VerifyType, req.AuthCode)
	if err != nil {
		return nil, errors.New("验证码错误")
	}
	return user, nil
}

// 使用第三方登陆时，第一次登陆默认同时注册用户
func (s *userServiceImpl) handleAuthByThirdParty(ctx context.Context, req *v1.LoginRequest) (*model.User, error) {
	var (
		user *model.User
		err  error
	)
	// 获取第三方用户信息
	thirdPartyAdapter, ok := service.GetThirdPartyAdapter(req.VerifyType)
	if !ok {
		return nil, errors.New("不支持的第三方登录类型")
	}
	thirdPartyUser, err := thirdPartyAdapter.FetchUserInfo(ctx, req.AccessToken, req.AuthCode)
	if err != nil {
		return nil, err
	}
	var thirdPartyMeta *model.ThirdPartyUserMeta

	// 1. 首先根据OpenID和提供商查找关联
	if s.thirdPartyUserDao != nil {
		thirdPartyMeta, err = s.thirdPartyUserDao.FindByOpenIDAndProvider(thirdPartyUser.OpenID, req.VerifyType)
		if err == nil && thirdPartyMeta != nil {
			// 找到关联，获取用户信息
			user, err = s.userDao.FindByID(thirdPartyMeta.UserID)
		}
		// 2. 如果OpenID未找到，尝试通过UnionID查找
		if thirdPartyUser.UnionID != "" && user == nil {
			thirdPartyMeta, err = s.thirdPartyUserDao.FindByUnionID(thirdPartyUser.UnionID)
			if err == nil && thirdPartyMeta != nil {
				// 找到关联，获取用户信息
				user, err = s.userDao.FindByID(thirdPartyMeta.UserID)
			}
		}
	}

	// 3. 如果没有找到关联，尝试通过邮箱查找用户
	if thirdPartyUser.Email != "" && user == nil {
		user, err = s.userDao.FindByEmail(thirdPartyUser.Email)
	}
	if thirdPartyUser.Phone != "" && user == nil {
		user, err = s.userDao.FindByPhone(thirdPartyUser.Phone)
	}
	if user == nil {
		// 4. 如果没有找到任何关联，创建新用户
		// 创建新用户
		user, err = s.createUser(req.VerifyType, req.Identifier, "", req.DeviceFingerprint, req.UserAgent, req.IP)
		if err != nil {
			return nil, errors.New("创建用户失败")
		}
	}
	err = s.linkThirdPartyUser(ctx, user.ID, thirdPartyUser)
	if err != nil {
		// TODO 创建关联失败不影响登录，记录日志
	}

	// 完成登录
	return user, nil
}

func (s *userServiceImpl) linkThirdPartyUser(ctx context.Context, userID uint, thirdPartyUser *model.ThirdPartyUser) error {
	if s.thirdPartyUserDao == nil {
		return errors.New("第三方用户仓库未初始化")
	}
	newMeta := &model.ThirdPartyUserMeta{
		UserID:      userID,
		Provider:    thirdPartyUser.Provider,
		OpenID:      thirdPartyUser.OpenID,
		UnionID:     thirdPartyUser.UnionID,
		AccessToken: thirdPartyUser.AccessToken,
	}
	_, err := s.thirdPartyUserDao.Create(newMeta)
	if err != nil {
		return err
	}
	return nil
}

// TODO 有内容需要完善
func (s *userServiceImpl) createUser(verifyType string, identifier string, password string, deviceFingerprint string, userAgent string, ip string) (*model.User, error) {
	var (
		user = &model.User{}
		err  error
	)
	if password == "" {
		password = generate.MustGenerateStrongPassword(16)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	switch verifyType {
	case "EMAIL":
		user.Email = identifier
	case "PHONE":
		user.Phone = identifier
	case "USERNAME":
		user.Username = identifier
	}
	if user.Username == "" {
		user.Username = "后续使用方法生成"
	}
	user.Nickname = user.Username
	user.Password = string(hashedPassword)
	user, err = s.userDao.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// LoginRecord 登录审计，记录登录历史和更新信任设备，由SSO回调
func (s *userServiceImpl) LoginRecord(ctx context.Context, req *model.LoginRecordRequest) {
	location := s.getLocationByIP(ctx, req.IP)
	loginHistory := &model.LoginHistory{
		UserID:            req.UserID,
		IP:                req.IP,
		DeviceFingerprint: req.DeviceFingerprint,
		UserAgent:         req.UserAgent,
		LoginTime:         time.Now(),
		Status:            req.Status,
		Device:            req.Device,
	}
	if s.loginHistoryDao != nil {
		s.loginHistoryDao.Create(loginHistory)
	}
	s.markTrustedDevice(req.UserID, req.Device, req.DeviceFingerprint, req.IP, location)

}

func (s *userServiceImpl) getLocationByIP(ctx context.Context, ip string) string {
	//TODO 根据IP获取位置信息
	return ""
}

// markTrustedDevice 标记信任设备
func (s *userServiceImpl) markTrustedDevice(userID uint, deviceName, deviceFingerprint string, ip, location string) {
	if s.trustedDeviceDao == nil || deviceFingerprint == "" {
		return
	}

	// 查找是否已存在信任设备
	device, err := s.trustedDeviceDao.FindByUserIDAndFingerprint(userID, deviceFingerprint)
	if err != nil {
		// 设备不存在，创建新的信任设备
		newDevice := &model.TrustedDevice{
			UserID:            userID,
			DeviceFingerprint: deviceFingerprint,
			DeviceName:        deviceName,
			LastLoginIP:       ip,
			FirstLoginTime:    time.Now(),
			LastLoginTime:     time.Now(),
			LastLoginLocation: location,
		}
		s.trustedDeviceDao.Create(newDevice)
	} else {
		// 设备已存在，更新最后登录时间和IP
		device.LastLoginTime = time.Now()
		device.LastLoginIP = ip
		device.LastLoginLocation = location
		s.trustedDeviceDao.Update(device)
	}
}

// detectTrustedDevice 检测设备是否已信任
func (s *userServiceImpl) detectTrustedDevice(userID uint, deviceFingerprint string) bool {
	if s.trustedDeviceDao == nil || deviceFingerprint == "" {
		return false
	}

	// 查找是否已存在信任设备
	_, err := s.trustedDeviceDao.FindByUserIDAndFingerprint(userID, deviceFingerprint)
	if err != nil {
		return false
	}

	return true
}

// SendVerifyCode 发送验证码 todo 完善发送验证码逻辑
func (s *userServiceImpl) SendVerifyCode(req *model.SendVerifyCodeRequest) error {
	ctx := context.Background()

	// 生成6位数字验证码
	verifyCode := generate.MustCaptcha(6)

	// 构建Redis键
	var key string
	switch req.VerifyType {
	case "EMAIL":
		key = fmt.Sprintf("verify_code:email:%s:%s", req.Identifier, req.Action)
	case "PHONE":
		key = fmt.Sprintf("verify_code:phone:%s:%s", req.Identifier, req.Action)
	default:
		return errors.New("无效的验证类型")
	}
	// TODO 不应该强依赖Redis，使用数据库存储，当Redis不可用时，应该使用数据库存储验证码
	// 存储验证码到Redis，设置5分钟过期
	redisClient := db.GetRedis()
	if redisClient == nil {
		logger.Errorf(ctx, "Redis客户端未初始化")
		return errors.New("验证码发送失败")
	}

	err := redisClient.Set(ctx, key, verifyCode, 5*time.Minute).Err()
	if err != nil {
		logger.Errorf(ctx, "存储验证码到Redis失败: %v", err)
		return errors.New("验证码发送失败")
	}

	// 发送验证码
	switch req.VerifyType {
	case "EMAIL":
		// 发送邮件验证码
		err = s.sendMail(req.Identifier, "验证码", fmt.Sprintf("您的验证码是: %s，5分钟内有效", verifyCode))
	case "PHONE":
		// 发送短信验证码
		err = s.sendSMS(req.Identifier, verifyCode)
	default:
		return errors.New("无效的验证类型")
	}

	if err != nil {
		logger.Errorf(ctx, "发送验证码失败: %v", err)
		return errors.New("验证码发送失败")
	}

	return nil
}

func (s *userServiceImpl) checkPassword(user *model.User, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return errors.New("密码错误")
	}
	return nil
}

// checkVerifyCode 验证验证码
func (s *userServiceImpl) checkVerifyCode(user *model.User, verifyCode string, verifyType string) error {
	ctx := context.Background()

	// 验证验证码
	var key string
	switch verifyType {
	case "EMAIL":
		key = fmt.Sprintf("verify_code:email:%s:login", user.Email)
	case "PHONE":
		key = fmt.Sprintf("verify_code:phone:%s:login", user.Phone)
	default:
		return errors.New("无效的验证类型")
	}

	// 从Redis获取验证码
	// TODO 不应该强依赖Redis，使用数据库存储，当Redis不可用时，应该使用数据库存储验证码
	redisClient := db.GetRedis()
	if redisClient == nil {
		logger.Errorf(ctx, "Redis客户端未初始化")
		return errors.New("验证码验证失败")
	}

	storedCode, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New("验证码已过期或不存在")
		}
		logger.Errorf(ctx, "从Redis获取验证码失败: %v", err)
		return errors.New("验证码验证失败")
	}

	// 验证验证码
	if storedCode != verifyCode {
		return errors.New("验证码错误")
	}

	// 验证码验证成功后删除
	_ = redisClient.Del(ctx, key).Err()

	return nil
}

// GetUserInfo 获取用户信息 todo 完善用户信息返回逻辑
func (s *userServiceImpl) GetUserInfo(userID uint) (*model.UserResponse, error) {
	// 查找用户
	user, err := s.userDao.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 转换为响应对象
	return user.ToResponse(), nil
}

// UpdateUserInfo 更新用户信息 todo 完善更新用户信息逻辑
func (s *userServiceImpl) UpdateUserInfo(userID uint, nickname string, phone string, avatar string) (*model.UserResponse, error) {
	// 查找用户
	user, err := s.userDao.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 更新用户信息
	if nickname != "" {
		user.Nickname = nickname
	}
	if phone != "" {
		user.Phone = phone
	}
	if avatar != "" {
		user.Avatar = avatar
	}

	// 保存更新
	user, err = s.userDao.Update(user)
	if err != nil {
		return nil, errors.New("更新用户信息失败")
	}

	// 转换为响应对象
	return user.ToResponse(), nil
}

// ChangePassword 修改密码 todo 完善修改密码逻辑
func (s *userServiceImpl) ChangePassword(userID uint, oldPassword string, newPassword string) error {
	// 查找用户
	user, err := s.userDao.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	user.Password = string(hashedPassword)

	// 保存更新
	_, err = s.userDao.Update(user)
	if err != nil {
		return errors.New("更新密码失败")
	}

	return nil
}

// Register 用户注册
func (s *userServiceImpl) Register(ctx context.Context, req *v1.RegisterRequest) (res *v1.RegisterResponse, err error) {
	if req.AuthType == "VERIFY_CODE" {
		// 验证验证码
		err = s.checkRegisterVerifyCode(req.Identifier, req.VerifyType, req.AuthCode)
		if err != nil {
			return nil, err
		}
	}

	// 检查用户是否已存在
	var existingUser *model.User
	switch req.VerifyType {
	case "USERNAME":
		existingUser, err = s.userDao.FindByUsername(req.Identifier)
	case "EMAIL":
		existingUser, err = s.userDao.FindByEmail(req.Identifier)
	case "PHONE":
		existingUser, err = s.userDao.FindByPhone(req.Identifier)
	default:
		return nil, errors.New("无效的验证类型")
	}
	if err == nil && existingUser != nil {
		return nil, errors.New("用户已存在")
	}

	// 创建用户
	user, err := s.createUser(req.VerifyType, req.Identifier, req.Password, req.DeviceFingerprint, req.UserAgent, req.IP)
	if err != nil {
		return nil, err
	}

	// 标记信任设备
	s.markTrustedDevice(user.ID, "", req.DeviceFingerprint, req.IP, "")

	// 返回注册结果
	return &v1.RegisterResponse{
		UserID: user.ID,
	}, nil
}

// checkRegisterVerifyCode 验证注册验证码
func (s *userServiceImpl) checkRegisterVerifyCode(identifier string, verifyType string, verifyCode string) error {
	ctx := context.Background()

	// 构建Redis键
	var key string
	switch verifyType {
	case "EMAIL":
		key = fmt.Sprintf("verify_code:email:%s:register", identifier)
	case "PHONE":
		key = fmt.Sprintf("verify_code:phone:%s:register", identifier)
	default:
		return errors.New("无效的验证类型")
	}

	// 从Redis获取验证码
	redisClient := db.GetRedis()
	if redisClient == nil {
		logger.Errorf(ctx, "Redis客户端未初始化")
		return errors.New("验证码验证失败")
	}

	storedCode, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New("验证码已过期或不存在")
		}
		logger.Errorf(ctx, "从Redis获取验证码失败: %v", err)
		return errors.New("验证码验证失败")
	}

	// 验证验证码
	if storedCode != verifyCode {
		return errors.New("验证码错误")
	}

	// 验证码验证成功后删除
	_ = redisClient.Del(ctx, key).Err()

	return nil
}

// ResetPassword 重置密码 todo 完善重置密码逻辑
func (s *userServiceImpl) ResetPassword(email string, verifyCode string, newPassword string) error {
	ctx := context.Background()

	// 验证验证码
	key := fmt.Sprintf("verify_code:email:%s:reset", email)

	// 从Redis获取验证码
	redisClient := db.GetRedis()
	if redisClient == nil {
		logger.Errorf(ctx, "Redis客户端未初始化")
		return errors.New("验证码验证失败")
	}

	storedCode, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New("验证码已过期或不存在")
		}
		logger.Errorf(ctx, "从Redis获取验证码失败: %v", err)
		return errors.New("验证码验证失败")
	}

	// 验证验证码
	if storedCode != verifyCode {
		return errors.New("验证码错误")
	}

	// 查找用户
	user, err := s.userDao.FindByEmail(email)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	user.Password = string(hashedPassword)

	// 保存更新
	_, err = s.userDao.Update(user)
	if err != nil {
		return errors.New("重置密码失败")
	}

	// 清除验证码
	_ = redisClient.Del(ctx, key).Err()

	return nil
}

// SendResetPasswordCode 发送重置密码验证码 todo 完善发送重置密码验证码逻辑
func (s *userServiceImpl) SendResetPasswordCode(email string) error {
	ctx := context.Background()

	// 检查用户是否存在
	_, err := s.userDao.FindByEmail(email)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 生成6位数字验证码
	verifyCode := generate.MustCaptcha(6)

	// 构建Redis键
	key := fmt.Sprintf("verify_code:email:%s:reset", email)

	// 存储验证码到Redis，设置5分钟过期
	redisClient := db.GetRedis()
	if redisClient == nil {
		logger.Errorf(ctx, "Redis客户端未初始化")
		return errors.New("验证码发送失败")
	}

	err = redisClient.Set(ctx, key, verifyCode, 5*time.Minute).Err()
	if err != nil {
		logger.Errorf(ctx, "存储验证码到Redis失败: %v", err)
		return errors.New("验证码发送失败")
	}

	// 发送验证码邮件
	err = s.sendMail(email, "重置密码验证码", fmt.Sprintf("您的重置密码验证码是: %s，5分钟内有效", verifyCode))
	if err != nil {
		logger.Errorf(ctx, "发送验证码失败: %v", err)
		return errors.New("验证码发送失败")
	}

	return nil
}

// detectDifferentLocation 检测异地登录
func (s *userServiceImpl) detectDifferentLocation(userID uint, ip string) bool {
	// TODO: 这里可以添加更复杂的异地登录检测逻辑
	// 例如：检查IP地址是否在常用地区范围内
	// 暂时简单返回true，表示异地登录
	// 注意如果用户没有历史登陆记录，为避免错误处理用户请求，需返回false
	return true
}

// sendMail 发送邮件
func (s *userServiceImpl) sendMail(to string, subject string, body string) error {
	mailClient := client.GetMailClient()
	if mailClient == nil {
		return errors.New("邮件客户端未初始化")
	}
	// 构建消息
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", s.config.Email.From, to, subject, body)

	// 发送邮件
	return mailClient.Send(&mail.Email{
		From:    s.config.Email.From,
		To:      []string{to},
		Subject: subject,
		Body:    message,
	})
}

// sendSMS 发送短信
func (s *userServiceImpl) sendSMS(phone string, code string) error {
	// 检查配置
	if s.config == nil || s.config.SMS.ApiKey == "" {
		return errors.New("短信配置未设置")
	}

	// 这里应该调用短信服务提供商的API
	// 暂时返回成功
	return nil
}
