package user

import (
	"context"
	"errors"
	"time"

	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/tools/generate"

	"golang.org/x/crypto/bcrypt"

	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/service"
)

func init() {
	di.Provide[service.UserService]("user.service", NewUserServiceImpl())
}

type userServiceImpl struct{}

func NewUserServiceImpl() *userServiceImpl {
	return &userServiceImpl{}
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
	if user.Status != 1 {
		return nil, errors.New("用户已禁用")
	}

	isNewDevice := s.detectTrustedDevice(user.ID, req.DeviceFingerprint)
	isDifferentLocation := s.detectDifferentLocation(user.ID, req.IP)

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
		user, err = dao.GetUserDao().FindByEmailFromDB(req.Identifier)
	case "PHONE":
		user, err = dao.GetUserDao().FindByPhoneFromDB(req.Identifier)
	case "USERNAME":
		user, err = dao.GetUserDao().FindByUsernameFromDB(req.Identifier)
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

func (s *userServiceImpl) handleAuthByVerifyCode(ctx context.Context, req *v1.LoginRequest) (*model.User, error) {
	var (
		user *model.User
		err  error
	)
	switch req.VerifyType {
	case "EMAIL":
		user, err = dao.GetUserDao().FindByEmailFromDB(req.Identifier)
	case "PHONE":
		user, err = dao.GetUserDao().FindByPhoneFromDB(req.Identifier)
	default:
		return nil, errors.New("无效的验证类型")
	}
	if err != nil {
		user, err = s.createUser(req.VerifyType, req.Identifier, "", req.DeviceFingerprint, req.UserAgent, req.IP)
		if err != nil {
			return nil, errors.New("用户不存在")
		}
	}
	commonService := service.GetCommonService()
	err = commonService.CheckVerifyCode(ctx, req.Identifier, req.VerifyType, req.AuthCode, "login")
	if err != nil {
		return nil, errors.New("验证码错误")
	}
	return user, nil
}

func (s *userServiceImpl) handleAuthByThirdParty(ctx context.Context, req *v1.LoginRequest) (*model.User, error) {
	var (
		user *model.User
		err  error
	)
	thirdPartyAdapter, ok := service.GetThirdPartyAdapter(req.VerifyType)
	if !ok {
		return nil, errors.New("不支持的第三方登录类型")
	}
	thirdPartyUser, err := thirdPartyAdapter.FetchUserInfo(ctx, req.AccessToken, req.AuthCode)
	if err != nil {
		return nil, err
	}
	var thirdPartyMeta *model.ThirdPartyUserMeta

	thirdPartyMeta, err = dao.GetThirdPartyUserDao().FindByOpenIDAndProvider(thirdPartyUser.OpenID, req.VerifyType)
	if err == nil && thirdPartyMeta != nil {
		user, err = dao.GetUserDao().FindByID(thirdPartyMeta.UserID)
	}
	if thirdPartyUser.UnionID != "" && user == nil {
		thirdPartyMeta, err = dao.GetThirdPartyUserDao().FindByUnionID(thirdPartyUser.UnionID)
		if err == nil && thirdPartyMeta != nil {
			user, err = dao.GetUserDao().FindByID(thirdPartyMeta.UserID)
		}
	}
	if thirdPartyUser.Email != "" && user == nil {
		user, err = dao.GetUserDao().FindByEmail(thirdPartyUser.Email)
	}
	if thirdPartyUser.Phone != "" && user == nil {
		user, err = dao.GetUserDao().FindByPhone(thirdPartyUser.Phone)
	}
	if user == nil {
		user, err = s.createUser(req.VerifyType, req.Identifier, "", req.DeviceFingerprint, req.UserAgent, req.IP)
		if err != nil {
			return nil, errors.New("创建用户失败")
		}
	}
	if linkErr := s.linkThirdPartyUser(ctx, user.ID, thirdPartyUser); linkErr != nil {
		// 创建关联失败不影响登录
	}

	return user, nil
}

func (s *userServiceImpl) linkThirdPartyUser(ctx context.Context, userID uint, thirdPartyUser *model.ThirdPartyUser) error {
	newMeta := &model.ThirdPartyUserMeta{
		UserID:      userID,
		Provider:    thirdPartyUser.Provider,
		OpenID:      thirdPartyUser.OpenID,
		UnionID:     thirdPartyUser.UnionID,
		AccessToken: thirdPartyUser.AccessToken,
	}
	_, err := dao.GetThirdPartyUserDao().Create(newMeta)
	return err
}

func (s *userServiceImpl) createUser(verifyType string, identifier string, password string, deviceFingerprint string, userAgent string, ip string) (*model.User, error) {
	if password == "" {
		password = generate.MustGenerateStrongPassword(16)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{}
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
	user, err = dao.GetUserDao().Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

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
	_, _ = dao.GetLoginHistoryDao().Create(loginHistory)

	s.markTrustedDevice(req.UserID, req.Device, req.DeviceFingerprint, req.IP, location)
}

func (s *userServiceImpl) getLocationByIP(ctx context.Context, ip string) string {
	return ""
}

func (s *userServiceImpl) markTrustedDevice(userID uint, deviceName, deviceFingerprint string, ip, location string) {
	if deviceFingerprint == "" {
		return
	}

	device, err := dao.GetTrustedDeviceDao().FindByUserIDAndFingerprint(userID, deviceFingerprint)
	if err != nil {
		newDevice := &model.TrustedDevice{
			UserID:            userID,
			DeviceFingerprint: deviceFingerprint,
			DeviceName:        deviceName,
			LastLoginIP:       ip,
			FirstLoginTime:    time.Now(),
			LastLoginTime:     time.Now(),
			LastLoginLocation: location,
		}
		_, _ = dao.GetTrustedDeviceDao().Create(newDevice)
	} else {
		device.LastLoginTime = time.Now()
		device.LastLoginIP = ip
		device.LastLoginLocation = location
		_, _ = dao.GetTrustedDeviceDao().Update(device)
	}
}

func (s *userServiceImpl) detectTrustedDevice(userID uint, deviceFingerprint string) bool {
	if deviceFingerprint == "" {
		return false
	}
	_, err := dao.GetTrustedDeviceDao().FindByUserIDAndFingerprint(userID, deviceFingerprint)
	return err == nil
}

func (s *userServiceImpl) checkPassword(user *model.User, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return errors.New("密码错误")
	}
	return nil
}

func (s *userServiceImpl) GetUserInfo(userID uint) (*model.User, error) {
	user, err := dao.GetUserDao().FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return user, nil
}

func (s *userServiceImpl) UpdateUserInfo(userID uint, nickname string, phone string, avatar string) (*model.User, error) {
	user, err := dao.GetUserDao().FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if nickname != "" {
		user.Nickname = nickname
	}
	if phone != "" {
		user.Phone = phone
	}
	if avatar != "" {
		user.Avatar = avatar
	}

	user, err = dao.GetUserDao().Update(user)
	if err != nil {
		return nil, errors.New("更新用户信息失败")
	}

	return user, nil
}

func (s *userServiceImpl) ChangePassword(userID uint, oldPassword string, newPassword string) error {
	user, err := dao.GetUserDao().FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	user.Password = string(hashedPassword)
	_, err = dao.GetUserDao().Update(user)
	if err != nil {
		return errors.New("更新密码失败")
	}

	return nil
}

func (s *userServiceImpl) Register(ctx context.Context, req *v1.RegisterRequest) (res *v1.RegisterResponse, err error) {
	if req.AuthType == "VERIFY_CODE" {
		commonService := service.GetCommonService()
		if err := commonService.CheckVerifyCode(ctx, req.Identifier, req.VerifyType, req.AuthCode, "register"); err != nil {
			return nil, err
		}
	}

	var existingUser *model.User
	switch req.VerifyType {
	case "USERNAME":
		existingUser, err = dao.GetUserDao().FindByUsername(req.Identifier)
	case "EMAIL":
		existingUser, err = dao.GetUserDao().FindByEmail(req.Identifier)
	case "PHONE":
		existingUser, err = dao.GetUserDao().FindByPhone(req.Identifier)
	default:
		return nil, errors.New("无效的验证类型")
	}
	if err == nil && existingUser != nil {
		return nil, errors.New("用户已存在")
	}

	user, err := s.createUser(req.VerifyType, req.Identifier, req.Password, req.DeviceFingerprint, req.UserAgent, req.IP)
	if err != nil {
		return nil, err
	}

	s.markTrustedDevice(user.ID, "", req.DeviceFingerprint, req.IP, "")

	return &v1.RegisterResponse{
		UserID: user.ID,
	}, nil
}

func (s *userServiceImpl) ResetPassword(email string, verifyCode string, newPassword string) error {
	ctx := context.Background()

	commonService := service.GetCommonService()
	if err := commonService.CheckVerifyCode(ctx, email, "EMAIL", verifyCode, "reset"); err != nil {
		return err
	}

	user, err := dao.GetUserDao().FindByEmail(email)
	if err != nil {
		return errors.New("用户不存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	user.Password = string(hashedPassword)
	_, err = dao.GetUserDao().Update(user)
	if err != nil {
		return errors.New("重置密码失败")
	}

	return nil
}

func (s *userServiceImpl) SendResetPasswordCode(email string) error {
	ctx := context.Background()

	_, err := dao.GetUserDao().FindByEmail(email)
	if err != nil {
		return errors.New("用户不存在")
	}

	commonService := service.GetCommonService()
	req := &model.SendVerifyCodeRequest{
		Identifier: email,
		VerifyType: "EMAIL",
		Action:     "reset",
	}
	return commonService.SendVerifyCode(ctx, req)
}

func (s *userServiceImpl) detectDifferentLocation(userID uint, ip string) bool {
	return true
}
