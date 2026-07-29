package user

import (
	"context"
	"time"

	"github.com/LandcLi/landc-go/frame/pkg/di"
	"github.com/LandcLi/landc-go/tools/generate"

	"golang.org/x/crypto/bcrypt"

	v1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/dao"
	lumerrors "github.com/LandcLi/LandcUserManager/pkg/errors"
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/service"
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

// Login 密码验证
func (s *userServiceImpl) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	var (
		user *model.User
		err  error
	)
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
	case constants.VerifyTypeEmail:
		user, err = dao.GetUserDao().FindByEmailFromDB(req.Identifier)
	case constants.VerifyTypePhone:
		user, err = dao.GetUserDao().FindByPhoneFromDB(req.Identifier)
	case constants.VerifyTypeUsername:
		user, err = dao.GetUserDao().FindByUsernameFromDB(req.Identifier)
	default:
		return nil, lumerrors.ErrInvalidVerifyType
	}
	if err != nil {
		return nil, lumerrors.ErrUserNotFound
	}
	err = s.checkPassword(user, req.AuthCode)
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
		user, err = dao.GetUserDao().FindByEmailFromDB(req.Identifier)
	case constants.VerifyTypePhone:
		user, err = dao.GetUserDao().FindByPhoneFromDB(req.Identifier)
	default:
		return nil, lumerrors.ErrInvalidVerifyType
	}
	if err != nil {
		user, err = s.createUser(req.VerifyType, req.Identifier, "", req.DeviceFingerprint, req.UserAgent, req.IP)
		if err != nil {
			return nil, lumerrors.ErrUserNotFound
		}
	}
	commonService := service.GetCommonService()
	err = commonService.CheckVerifyCode(ctx, req.Identifier, req.VerifyType, req.AuthCode, "login")
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
	thirdPartyAdapter, ok := service.GetThirdPartyAdapter(req.VerifyType)
	if !ok {
		return nil, lumerrors.ErrThirdPartyNotSupported
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
			return nil, lumerrors.ErrCreateUserFailed
		}
	}
	if linkErr := s.linkThirdPartyUser(ctx, user.ID, thirdPartyUser); linkErr != nil {
		logger.Warnf(ctx, "link third party user failed: %v", linkErr)
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
	if _, err := dao.GetLoginHistoryDao().Create(loginHistory); err != nil {
		logger.Errorf(ctx, "failed to record login history: %v", err)
	}

	s.markTrustedDevice(req.UserID, req.Device, req.DeviceFingerprint, req.IP, location)
}

func (s *userServiceImpl) getLocationByIP(ctx context.Context, ip string) string {
	// TODO: implement IP geolocation (planned feature)
	// Currently returns empty string until geolocation service is integrated
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
		if _, err := dao.GetTrustedDeviceDao().Create(newDevice); err != nil {
			logger.Errorf(context.Background(), "failed to create trusted device: %v", err)
		}
	} else {
		device.LastLoginTime = time.Now()
		device.LastLoginIP = ip
		device.LastLoginLocation = location
		if _, err := dao.GetTrustedDeviceDao().Update(device); err != nil {
			logger.Errorf(context.Background(), "failed to update trusted device: %v", err)
		}
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
		return lumerrors.ErrPasswordIncorrect
	}
	return nil
}

func (s *userServiceImpl) GetUserInfo(userID uint) (*model.User, error) {
	user, err := dao.GetUserDao().FindByID(userID)
	if err != nil {
		return nil, lumerrors.ErrUserNotFound
	}
	return user, nil
}

func (s *userServiceImpl) UpdateUserInfo(userID uint, nickname string, phone string, avatar string) (*model.User, error) {
	user, err := dao.GetUserDao().FindByID(userID)
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

	user, err = dao.GetUserDao().Update(user)
	if err != nil {
		return nil, lumerrors.ErrUpdateUserFailed
	}

	return user, nil
}

func (s *userServiceImpl) ChangePassword(userID uint, oldPassword string, newPassword string) error {
	user, err := dao.GetUserDao().FindByID(userID)
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
	_, err = dao.GetUserDao().Update(user)
	if err != nil {
		return lumerrors.ErrUpdatePasswordFail
	}

	return nil
}

func (s *userServiceImpl) Register(ctx context.Context, req *v1.RegisterRequest) (res *v1.RegisterResponse, err error) {
	if req.AuthType == constants.AuthTypeVerifyCode {
		commonService := service.GetCommonService()
		if err := commonService.CheckVerifyCode(ctx, req.Identifier, req.VerifyType, req.AuthCode, "register"); err != nil {
			return nil, err
		}
	}

	var existingUser *model.User
	switch req.VerifyType {
	case constants.VerifyTypeUsername:
		existingUser, err = dao.GetUserDao().FindByUsername(req.Identifier)
	case constants.VerifyTypeEmail:
		existingUser, err = dao.GetUserDao().FindByEmail(req.Identifier)
	case constants.VerifyTypePhone:
		existingUser, err = dao.GetUserDao().FindByPhone(req.Identifier)
	default:
		return nil, lumerrors.ErrInvalidVerifyType
	}
	if err == nil && existingUser != nil {
		return nil, lumerrors.ErrUserAlreadyExists
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
	if err := commonService.CheckVerifyCode(ctx, email, constants.VerifyTypeEmail, verifyCode, "reset"); err != nil {
		return err
	}

	user, err := dao.GetUserDao().FindByEmail(email)
	if err != nil {
		return lumerrors.ErrUserNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return lumerrors.ErrPasswordHashFailed
	}

	user.Password = string(hashedPassword)
	_, err = dao.GetUserDao().Update(user)
	if err != nil {
		return lumerrors.ErrResetPasswordFail
	}

	return nil
}

func (s *userServiceImpl) SendResetPasswordCode(email string) error {
	ctx := context.Background()

	_, err := dao.GetUserDao().FindByEmail(email)
	if err != nil {
		return lumerrors.ErrUserNotFound
	}

	commonService := service.GetCommonService()
	req := &model.SendVerifyCodeRequest{
		Identifier: email,
		VerifyType: constants.VerifyTypeEmail,
		Action:     "reset",
	}
	return commonService.SendVerifyCode(ctx, req)
}

func (s *userServiceImpl) detectDifferentLocation(userID uint, ip string) bool {
	// TODO: implement IP geolocation detection (planned feature)
	// Currently returns true as a safe default until geolocation service is integrated
	return true
}
