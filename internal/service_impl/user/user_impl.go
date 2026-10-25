package user

import (
	"context"
	"errors"
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
	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/LandcLi/LandcUserManager/utility/crypto"
	"github.com/LandcLi/LandcUserManager/utility/logger"
	"github.com/LandcLi/LandcUserManager/utility/loginguard"

	constants "github.com/LandcLi/LandcUserManager/pkg/consts"
)

func init() {
	di.Provide[service.UserService]("user.service", NewUserServiceImpl())
}

type userServiceImpl struct{}

func NewUserServiceImpl() *userServiceImpl {
	return &userServiceImpl{}
}

// bcryptCost 读取 password.bcrypt_cost 配置。
//
// 此前三处生成密码哈希的地方都硬编码 bcrypt.DefaultCost，配置项形同虚设
// （生产环境调高 cost 不生效，测试环境调低 cost 也没法提速）。
// 这里按配置取值并夹到 bcrypt 的合法区间，非法配置不会让注册直接失败。
func bcryptCost() int {
	cost := bcrypt.DefaultCost
	if cfg := config.TryGetConfig(); cfg != nil && cfg.Password.BcryptCost > 0 {
		cost = cfg.Password.BcryptCost
	}
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}
	return cost
}

// Login 登录入口：认证成功后签发 JWT（access + refresh），并自动记录登录历史
func (s *userServiceImpl) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
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

	user, err := s.authenticate(ctx, req)
	if err != nil {
		return nil, err
	}
	// 状态校验放在认证之后：此时凭据已经验证通过，回 ErrUserDisabled
	// 不会泄露「账号是否存在」（攻击者得先知道密码）。
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
		Identifier:        loginguard.Normalize(req.Identifier),
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

// authenticate 按认证方式校验身份。
//
// 只有密码登录承担失败计数与账号锁定：
//   - 验证码登录本身已受 frame/pkg/verifycode 的发送间隔与每日上限约束，
//     且验证码是一次性、短时效的，再加一层时间锁只会把正常用户拦在门外；
//   - 第三方登录的凭据由 provider 校验，锁定发生在上游。
func (s *userServiceImpl) authenticate(ctx context.Context, req *v1.LoginRequest) (*model.User, error) {
	switch req.AuthType {
	case constants.AuthTypeVerifyCode:
		return s.handleAuthByVerifyCode(ctx, req)
	case constants.AuthTypeThirdParty:
		return s.handleAuthByThirdParty(ctx, req)
	case constants.AuthTypePassword:
		// 继续走下面的密码分支
	default:
		return nil, lumerrors.ErrInvalidAuthType
	}

	// 锁定期内直接拒绝：不查库、不比密码。既省资源，也避免用响应时间差
	// 推测「账号是否存在」（这是锁定之外的另一条枚举侧信道）。
	if locked, remaining := loginguard.Locked(ctx, req.Identifier); locked {
		return nil, lockedError(remaining)
	}

	user, err := s.handleAuthByPassword(ctx, req)
	if err != nil {
		// 只有「凭据不正确」参与计数。ErrInvalidVerifyType 之类的参数错误
		// 不计数，否则前端传错字段就能把用户锁死。
		if errors.Is(err, lumerrors.ErrInvalidCredentials) {
			var userID uint
			if user != nil {
				userID = user.ID
			}
			locked, remaining := loginguard.OnFailure(ctx, req.Identifier)
			if locked {
				s.recordFailedLogin(ctx, userID, req, "多次密码错误，账号临时锁定")
				return nil, lockedError(remaining)
			}
			s.recordFailedLogin(ctx, userID, req, "密码错误")
		}
		return nil, err
	}

	loginguard.OnSuccess(ctx, req.Identifier)
	return user, nil
}

// lockedError 把剩余锁定时间转成对外错误。
// remaining 非正数时按 1 分钟兜底，保证返回值永远是「锁定」而不是别的语义。
func lockedError(remaining time.Duration) error {
	if remaining <= 0 {
		remaining = time.Minute
	}
	return lumerrors.NewAccountLockedError(remaining)
}

// recordFailedLogin 记录一次失败登录（login_histories.status = 0）。
//
// 只在「真的校验失败」时写库。锁定期内被直接拒绝的请求**不落库** ——
// 否则攻击者持续发请求就能把登录历史表撑爆（放大攻击）。
// 有了成功路径的计数上限（最多 max_failed_attempts 条），每轮攻击的落库量是常数。
func (s *userServiceImpl) recordFailedLogin(ctx context.Context, userID uint, req *v1.LoginRequest, remark string) {
	history := &model.LoginHistory{
		UserID:            userID,
		Identifier:        loginguard.Normalize(req.Identifier),
		IP:                req.IP,
		DeviceFingerprint: req.DeviceFingerprint,
		UserAgent:         req.UserAgent,
		LoginTime:         time.Now(),
		Status:            0,
		Remark:            remark,
	}
	if _, err := dao.GetLoginHistoryDao().Create(ctx, history); err != nil {
		logger.Errorf(ctx, "记录失败登录历史失败: %v", err)
	}
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

// handleAuthByPassword 校验密码。
//
// 返回值语义：出错时如果已经查到用户，会把用户一并返回（err != nil && user != nil），
// 供调用方写失败登录记录；「查无此人」时 user 为 nil。
// 对外统一返回 ErrInvalidCredentials —— 「用户不存在」与「密码错误」必须不可区分，
// 否则可以用不同响应把用户名批量枚举出来。
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
	if err != nil || user == nil {
		// 不打日志到 warn 级别以上、也不带标识符，避免日志里堆积 PII；
		// 真正的取证信息在 login_histories.identifier 字段。
		logger.Infof(ctx, "密码登录：标识符未命中，verifyType=%s", req.VerifyType)
		return nil, lumerrors.ErrInvalidCredentials
	}
	if err := s.checkPassword(ctx, user, req.AuthCode); err != nil {
		return user, lumerrors.ErrInvalidCredentials
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
	if err != nil || user == nil {
		// 安全考虑：验证码登录不允许自动注册，防止验证码爆破自动批量开户。
		// 同样返回统一的凭据错误，不暴露账号是否存在。
		return nil, lumerrors.ErrInvalidCredentials
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost())
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
		Identifier:        req.Identifier,
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost())
	if err != nil {
		return lumerrors.ErrPasswordHashFailed
	}

	user.Password = string(hashedPassword)
	_, err = dao.GetUserDao().Update(ctx, user)
	if err != nil {
		return lumerrors.ErrUpdatePasswordFail
	}

	// 改密说明本人已证明身份，顺手解开因撞库导致的临时锁定，
	// 否则「被锁 + 改密」之后仍要干等一段时间才能登录。
	loginguard.Unlock(ctx, user.Username, user.Email, user.Phone)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost())
	if err != nil {
		return lumerrors.ErrPasswordHashFailed
	}

	user.Password = string(hashedPassword)
	_, err = dao.GetUserDao().Update(ctx, user)
	if err != nil {
		return lumerrors.ErrResetPasswordFail
	}

	// 重置密码同样视为「已证明身份」，解除临时锁定。
	loginguard.Unlock(ctx, user.Username, user.Email, user.Phone)
	return nil
}

func (s *userServiceImpl) SendResetPasswordCode(ctx context.Context, email string) error {
	user, err := dao.GetUserDao().FindByEmail(ctx, email)
	if err != nil || user == nil {
		// 防邮箱枚举：邮箱未注册时**不做任何提示**，直接返回成功。
		// 客户端统一展示「如果该邮箱已注册，验证码已发送」。
		// 后续 ResetPassword 会因为取不到验证码而失败，链路依然闭环。
		logger.Infof(ctx, "重置密码：邮箱未注册，静默跳过发送")
		return nil
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
