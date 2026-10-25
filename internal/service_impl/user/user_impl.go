package user

import (
	"context"
	"strings"
	"time"

	frameauth "github.com/LandcLi/landc-go/frame/pkg/auth"
	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/tools/generate"

	"golang.org/x/crypto/bcrypt"

	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/dao"
	"github.com/LandcLi/LandcUserManager/model"
	lumerrors "github.com/LandcLi/LandcUserManager/pkg/errors"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/LandcUserManager/utility/crypto"
	"github.com/LandcLi/LandcUserManager/utility/logger"

	constants "github.com/LandcLi/LandcUserManager/pkg/consts"
)

func init() {
	di.Provide[service.UserService]("user.service", NewUserServiceImpl())
}

type userServiceImpl struct{}

func NewUserServiceImpl() *userServiceImpl {
	return &userServiceImpl{}
}

// Login 登录入口：认证成功后签发 JWT（access + refresh），并自动记录登录历史
func (s *userServiceImpl) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	var (
		user *model.User
		err  error
	)
	// binding 已放开 required，此处按 AuthType 分支承担必填校验（防安全退化）
	switch req.AuthType {
	case constants.AuthTypePassword, constants.AuthTypeVerifyCode:
		if req.Identifier == "" || req.AuthCode == "" {
			return nil, lumerrors.ErrInvalidRequestParams
		}
	case constants.AuthTypeThirdParty:
		if req.AccessToken == "" {
			return nil, lumerrors.ErrInvalidRequestParams
		}
	}

	switch req.AuthType {
	case constants.AuthTypePassword:
		user, err = s.handleAuthByPassword(ctx, req)
	case constants.AuthTypeVerifyCode:
		user, err = s.handleAuthByVerifyCode(ctx, req)
	case constants.AuthTypeThirdParty:
		user, err = s.handleAuthByThirdParty(ctx, req)
	default:
		return nil, lumerrors.ErrInvalidAuthType
	}
	if err != nil {
		return nil, err
	}
	if user.Status != 1 {
		return nil, lumerrors.ErrUserDisabled
	}

	// 注意语义：detectTrustedDevice 返回「设备可信」，取反才是「是否新设备」。
	// 此前直接赋给 isNewDevice，导致可信设备被判为新设备、真陌生设备反而不触发
	// 临时令牌机制（SSO 的新设备防护因此失效）。
	isNewDevice := req.DeviceFingerprint != "" && !s.detectTrustedDevice(ctx, user.ID, req.DeviceFingerprint)
	isDifferentLocation := s.detectDifferentLocation(ctx, user.ID, req.IP)

	// 签发 JWT
	accessToken, refreshToken, err := s.issueTokens(ctx, user, req.IP, req.DeviceFingerprint)
	if err != nil {
		logger.Errorf(ctx, "签发 JWT 失败: %v", err)
		return nil, lumerrors.ErrUserNotFound
	}

	// 自动记录登录历史（含信任设备管理）
	s.LoginRecord(ctx, &model.LoginRecordRequest{
		UserID:            user.ID,
		IP:                req.IP,
		DeviceFingerprint: req.DeviceFingerprint,
		UserAgent:         req.UserAgent,
		Status:            1,
	})

	return &v1.LoginResponse{
		UserID:            user.ID,
		DifferentLocation: isDifferentLocation,
		IsNewDevice:       isNewDevice,
		Token:             accessToken,
		RefreshToken:      refreshToken,
	}, nil
}

// issueTokens 签发 access token 与 refresh token
func (s *userServiceImpl) issueTokens(ctx context.Context, user *model.User, ip, deviceFingerprint string) (string, string, error) {
	access, err := frameauth.GenerateTokenWithOpts(user.ID, user.Username, frameauth.GenerateTokenOption{
		IP:                ip,
		DeviceFingerprint: deviceFingerprint,
		TokenType:         "access",
	})
	if err != nil {
		return "", "", err
	}

	refresh, err := frameauth.GenerateTokenWithOpts(user.ID, user.Username, frameauth.GenerateTokenOption{
		IP:                ip,
		DeviceFingerprint: deviceFingerprint,
		TokenType:         "refresh",
		ExpireTime:        30 * 24 * time.Hour, // refresh token 有效期 30 天
	})
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (s *userServiceImpl) handleAuthByPassword(ctx context.Context, req *v1.LoginRequest) (*model.User, error) {
	var (
		user *model.User
		err  error
	)
	switch req.VerifyType {
	case constants.VerifyTypeEmail:
		user, err = dao.GetUserDao().FindByEmailFromDB(ctx, req.Identifier)
	case constants.VerifyTypePhone:
		user, err = dao.GetUserDao().FindByPhoneFromDB(ctx, req.Identifier)
	case constants.VerifyTypeUsername:
		user, err = dao.GetUserDao().FindByUsernameFromDB(ctx, req.Identifier)
	default:
		return nil, lumerrors.ErrInvalidVerifyType
	}
	if err != nil {
		return nil, lumerrors.ErrUserNotFound
	}
	err = s.checkPassword(ctx, user, req.AuthCode)
	if err != nil {
		return nil, lumerrors.ErrPasswordIncorrect
	}
	return user, nil
}

func (s *userServiceImpl) handleAuthByVerifyCode(ctx context.Context, req *v1.LoginRequest) (*model.User, error) {
	var (
		user *model.User
		err  error
	)
	switch req.VerifyType {
	case constants.VerifyTypeEmail:
		user, err = dao.GetUserDao().FindByEmailFromDB(ctx, req.Identifier)
	case constants.VerifyTypePhone:
		user, err = dao.GetUserDao().FindByPhoneFromDB(ctx, req.Identifier)
	default:
		return nil, lumerrors.ErrInvalidVerifyType
	}
	if err != nil {
		// 安全考虑：验证码登录不允许自动注册，防止验证码爆破自动批量开户
		return nil, lumerrors.ErrUserNotFound
	}
	commonService := service.GetCommonService()
	err = commonService.CheckVerifyCode(ctx, req.Identifier, req.VerifyType, req.AuthCode, constants.ActionLogin)
	if err != nil {
		return nil, lumerrors.ErrVerifyCodeWrong
	}
	return user, nil
}

func (s *userServiceImpl) handleAuthByThirdParty(ctx context.Context, req *v1.LoginRequest) (*model.User, error) {
	var (
		user *model.User
		err  error
	)
	thirdPartyAdapter, ok := service.GetThirdPartyAdapter(strings.ToLower(req.VerifyType))
	if !ok {
		return nil, lumerrors.ErrThirdPartyNotSupported
	}
	thirdPartyUser, err := thirdPartyAdapter.FetchUserInfo(ctx, req.AccessToken, req.AuthCode)
	if err != nil {
		return nil, err
	}
	var thirdPartyMeta *model.ThirdPartyUserMeta

	thirdPartyMeta, err = dao.GetThirdPartyUserDao().FindByOpenIDAndProvider(ctx, thirdPartyUser.OpenID, req.VerifyType)
	if err == nil && thirdPartyMeta != nil {
		user, err = dao.GetUserDao().FindByID(ctx, thirdPartyMeta.UserID)
	}
	if thirdPartyUser.UnionID != "" && user == nil {
		thirdPartyMeta, err = dao.GetThirdPartyUserDao().FindByUnionID(ctx, thirdPartyUser.UnionID)
		if err == nil && thirdPartyMeta != nil {
			user, err = dao.GetUserDao().FindByID(ctx, thirdPartyMeta.UserID)
		}
	}
	if thirdPartyUser.Email != "" && user == nil {
		user, err = dao.GetUserDao().FindByEmail(ctx, thirdPartyUser.Email)
	}
	if thirdPartyUser.Phone != "" && user == nil {
		user, err = dao.GetUserDao().FindByPhone(ctx, thirdPartyUser.Phone)
	}
	if user == nil {
		// 第三方登录：首次登录自动创建用户（三方登录属于明确的用户授权行为）。
		// 优先用第三方返回的邮箱/手机建号（否则 Identifier 为空时退化为随机用户名）。
		createType, createIdentifier := req.VerifyType, ""
		switch {
		case thirdPartyUser.Email != "":
			createType, createIdentifier = constants.VerifyTypeEmail, thirdPartyUser.Email
		case thirdPartyUser.Phone != "":
			createType, createIdentifier = constants.VerifyTypePhone, thirdPartyUser.Phone
		}
		user, err = s.createUser(ctx, createType, createIdentifier, "", req.DeviceFingerprint, req.UserAgent, req.IP)
		if err != nil {
			return nil, lumerrors.ErrCreateUserFailed
		}
	}
	if linkErr := s.linkThirdPartyUser(ctx, user.ID, thirdPartyUser); linkErr != nil {
		logger.Warnf(ctx, "link third party user failed: %v", linkErr)
	}

	return user, nil
}

func (s *userServiceImpl) linkThirdPartyUser(ctx context.Context, userID uint, thirdPartyUser *model.ThirdPartyUser) error {
	// AccessToken 加密后存储，避免明文泄露
	encryptedToken, err := crypto.Encrypt(thirdPartyUser.AccessToken)
	if err != nil {
		logger.Warnf(ctx, "encrypt third party access token failed: %v", err)
		encryptedToken = thirdPartyUser.AccessToken
	}

	// Upsert：已存在绑定则更新 token，否则新建，避免重复 INSERT 与 token 过期后无法刷新
	existing, err := dao.GetThirdPartyUserDao().FindByOpenIDAndProvider(ctx, thirdPartyUser.OpenID, thirdPartyUser.Provider)
	if err == nil && existing != nil {
		existing.AccessToken = encryptedToken
		_, err = dao.GetThirdPartyUserDao().Update(ctx, existing)
		return err
	}

	newMeta := &model.ThirdPartyUserMeta{
		UserID:      userID,
		Provider:    thirdPartyUser.Provider,
		OpenID:      thirdPartyUser.OpenID,
		UnionID:     thirdPartyUser.UnionID,
		AccessToken: encryptedToken,
	}
	_, err = dao.GetThirdPartyUserDao().Create(ctx, newMeta)
	return err
}

func (s *userServiceImpl) createUser(ctx context.Context, verifyType string, identifier string, password string, deviceFingerprint string, userAgent string, ip string) (*model.User, error) {
	if password == "" {
		password = generate.MustGenerateStrongPassword(16)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{}
	switch verifyType {
	case constants.VerifyTypeEmail:
		user.Email = identifier
	case constants.VerifyTypePhone:
		user.Phone = identifier
	case constants.VerifyTypeUsername:
		user.Username = identifier
	}
	if user.Username == "" {
		user.Username = "user_" + generate.UUID()
	}
	user.Nickname = user.Username
	user.Password = string(hashedPassword)
	user, err = dao.GetUserDao().Create(ctx, user)
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
	if _, err := dao.GetLoginHistoryDao().Create(ctx, loginHistory); err != nil {
		logger.Errorf(ctx, "failed to record login history: %v", err)
	}

	s.markTrustedDevice(ctx, req.UserID, req.Device, req.DeviceFingerprint, req.IP, location)
}

func (s *userServiceImpl) getLocationByIP(ctx context.Context, ip string) string {
	// TODO: implement IP geolocation (planned feature)
	// Currently returns empty string until geolocation service is integrated
	return ""
}

func (s *userServiceImpl) markTrustedDevice(ctx context.Context, userID uint, deviceName, deviceFingerprint string, ip, location string) {
	if deviceFingerprint == "" {
		return
	}

	device, err := dao.GetTrustedDeviceDao().FindByUserIDAndFingerprint(ctx, userID, deviceFingerprint)
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
		if _, err := dao.GetTrustedDeviceDao().Create(ctx, newDevice); err != nil {
			logger.Errorf(ctx, "failed to create trusted device: %v", err)
		}
	} else {
		device.LastLoginTime = time.Now()
		device.LastLoginIP = ip
		device.LastLoginLocation = location
		if _, err := dao.GetTrustedDeviceDao().Update(ctx, device); err != nil {
			logger.Errorf(ctx, "failed to update trusted device: %v", err)
		}
	}
}

func (s *userServiceImpl) detectTrustedDevice(ctx context.Context, userID uint, deviceFingerprint string) bool {
	if deviceFingerprint == "" {
		return false
	}
	_, err := dao.GetTrustedDeviceDao().FindByUserIDAndFingerprint(ctx, userID, deviceFingerprint)
	return err == nil
}

func (s *userServiceImpl) checkPassword(ctx context.Context, user *model.User, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return lumerrors.ErrPasswordIncorrect
	}
	return nil
}

func (s *userServiceImpl) GetUserInfo(ctx context.Context, userID uint) (*model.User, error) {
	user, err := dao.GetUserDao().FindByID(ctx, userID)
	if err != nil {
		return nil, lumerrors.ErrUserNotFound
	}
	return user, nil
}

func (s *userServiceImpl) UpdateUserInfo(ctx context.Context, userID uint, nickname string, phone string, avatar string) (*model.User, error) {
	user, err := dao.GetUserDao().FindByID(ctx, userID)
	if err != nil {
		return nil, lumerrors.ErrUserNotFound
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

	user, err = dao.GetUserDao().Update(ctx, user)
	if err != nil {
		return nil, lumerrors.ErrUpdateUserFailed
	}

	return user, nil
}

func (s *userServiceImpl) ChangePassword(ctx context.Context, userID uint, oldPassword string, newPassword string) error {
	user, err := dao.GetUserDao().FindByID(ctx, userID)
	if err != nil {
		return lumerrors.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return lumerrors.ErrOldPasswordWrong
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return lumerrors.ErrPasswordHashFailed
	}

	user.Password = string(hashedPassword)
	_, err = dao.GetUserDao().Update(ctx, user)
	if err != nil {
		return lumerrors.ErrUpdatePasswordFail
	}

	return nil
}

func (s *userServiceImpl) Register(ctx context.Context, req *v1.RegisterRequest) (res *v1.RegisterResponse, err error) {
	if req.AuthType == constants.AuthTypeVerifyCode {
		commonService := service.GetCommonService()
		if err := commonService.CheckVerifyCode(ctx, req.Identifier, req.VerifyType, req.AuthCode, constants.ActionRegister); err != nil {
			return nil, err
		}
	}

	var existingUser *model.User
	switch req.VerifyType {
	case constants.VerifyTypeUsername:
		existingUser, err = dao.GetUserDao().FindByUsername(ctx, req.Identifier)
	case constants.VerifyTypeEmail:
		existingUser, err = dao.GetUserDao().FindByEmail(ctx, req.Identifier)
	case constants.VerifyTypePhone:
		existingUser, err = dao.GetUserDao().FindByPhone(ctx, req.Identifier)
	default:
		return nil, lumerrors.ErrInvalidVerifyType
	}
	if err == nil && existingUser != nil {
		return nil, lumerrors.ErrUserAlreadyExists
	}

	user, err := s.createUser(ctx, req.VerifyType, req.Identifier, req.Password, req.DeviceFingerprint, req.UserAgent, req.IP)
	if err != nil {
		return nil, err
	}

	s.markTrustedDevice(ctx, user.ID, "", req.DeviceFingerprint, req.IP, "")

	// 注册成功后直接签发 token，完成认证闭环
	accessToken, refreshToken, err := s.issueTokens(ctx, user, req.IP, req.DeviceFingerprint)
	if err != nil {
		logger.Errorf(ctx, "签发 JWT 失败: %v", err)
		return nil, lumerrors.ErrRegisterTokenFailed
	}

	return &v1.RegisterResponse{
		UserID:       user.ID,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *userServiceImpl) ResetPassword(ctx context.Context, email string, verifyCode string, newPassword string) error {
	commonService := service.GetCommonService()
	if err := commonService.CheckVerifyCode(ctx, email, constants.VerifyTypeEmail, verifyCode, constants.ActionReset); err != nil {
		return err
	}

	user, err := dao.GetUserDao().FindByEmail(ctx, email)
	if err != nil {
		return lumerrors.ErrUserNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return lumerrors.ErrPasswordHashFailed
	}

	user.Password = string(hashedPassword)
	_, err = dao.GetUserDao().Update(ctx, user)
	if err != nil {
		return lumerrors.ErrResetPasswordFail
	}

	return nil
}

func (s *userServiceImpl) SendResetPasswordCode(ctx context.Context, email string) error {
	_, err := dao.GetUserDao().FindByEmail(ctx, email)
	if err != nil {
		return lumerrors.ErrUserNotFound
	}

	commonService := service.GetCommonService()
	req := &model.SendVerifyCodeRequest{
		Identifier: email,
		VerifyType: constants.VerifyTypeEmail,
		Action:     constants.ActionReset,
	}
	return commonService.SendVerifyCode(ctx, req)
}

// detectDifferentLocation 判断本次登录是否为「异地」。
//
// 说明：这里只能做 IP 层面的粗判（与最近一次成功登录的 IP 不同即视为异地），
// 真正的地理位置解析属于风控范畴，需要 IP 归属地服务，见 TODO。
// 此前硬编码 return true，会让每次登录都被标记为异地，风控信号彻底失效。
func (s *userServiceImpl) detectDifferentLocation(ctx context.Context, userID uint, ip string) bool {
	if ip == "" {
		return false
	}
	history, err := dao.GetLoginHistoryDao().FindLatestByUserID(ctx, userID)
	if err != nil || history == nil || history.IP == "" {
		// 无历史记录（首次登录）不判定为异地
		return false
	}
	return history.IP != ip
}
