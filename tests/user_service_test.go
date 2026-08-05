package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/model"
	lumerrors "github.com/LandcLi/LandcUserManager/pkg/errors"
	constants "github.com/LandcLi/LandcUserManager/pkg/consts"
	"golang.org/x/crypto/bcrypt"
)

// ========= 辅助函数 =========

func hashPassword(pwd string) string {
	b, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(b)
}

func addTestUser(username, email, phone, password string, status int) *model.User {
	u := &model.User{
		Username: username,
		Email:    email,
		Phone:    phone,
		Password: hashPassword(password),
		Nickname: username,
		Status:   status,
	}
	mockUserDao.AddUser(u)
	return u
}

// ========= Login - PASSWORD 测试 =========

func TestLoginPasswordByUsername_Success(t *testing.T) {
	reset()
	addTestUser("testuser", "test@example.com", "", "pass123", 1)

	resp, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "testuser",
		AuthCode:   "pass123",
		AuthType:   constants.AuthTypePassword,
		VerifyType: constants.VerifyTypeUsername,
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.UserID == 0 {
		t.Fatal("expected non-zero UserID")
	}
	if resp.Token == "" {
		t.Fatal("expected JWT token in login response")
	}
	if resp.RefreshToken == "" {
		t.Fatal("expected refresh token in login response")
	}
}

func TestLoginPasswordByEmail_Success(t *testing.T) {
	reset()
	addTestUser("user1", "user1@example.com", "", "pwd456", 1)

	resp, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "user1@example.com",
		AuthCode:   "pwd456",
		AuthType:   constants.AuthTypePassword,
		VerifyType: constants.VerifyTypeEmail,
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.UserID == 0 {
		t.Fatal("expected non-zero UserID")
	}
}

func TestLoginPasswordByPhone_Success(t *testing.T) {
	reset()
	addTestUser("user2", "", "13800138000", "pwd789", 1)

	resp, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "13800138000",
		AuthCode:   "pwd789",
		AuthType:   constants.AuthTypePassword,
		VerifyType: constants.VerifyTypePhone,
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.UserID == 0 {
		t.Fatal("expected non-zero UserID")
	}
}

func TestLoginPassword_UserNotFound(t *testing.T) {
	reset()
	_, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "nonexistent",
		AuthCode:   "pass123",
		AuthType:   constants.AuthTypePassword,
		VerifyType: constants.VerifyTypeUsername,
	})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	if !errors.Is(err, lumerrors.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestLoginPassword_WrongPassword(t *testing.T) {
	reset()
	addTestUser("user3", "", "", "correctPassword", 1)

	_, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "user3",
		AuthCode:   "wrongPassword",
		AuthType:   constants.AuthTypePassword,
		VerifyType: constants.VerifyTypeUsername,
	})
	if err == nil || !errors.Is(err, lumerrors.ErrPasswordIncorrect) {
		t.Fatalf("expected ErrPasswordIncorrect error, got: %v", err)
	}
}

func TestLoginPassword_DisabledUser(t *testing.T) {
	reset()
	addTestUser("disabledUser", "", "", "pass123", 0)

	_, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "disabledUser",
		AuthCode:   "pass123",
		AuthType:   constants.AuthTypePassword,
		VerifyType: constants.VerifyTypeUsername,
	})
	if err == nil || !errors.Is(err, lumerrors.ErrUserDisabled) {
		t.Fatalf("expected ErrUserDisabled error, got: %v", err)
	}
}

func TestLoginPassword_InvalidVerifyType(t *testing.T) {
	reset()
	_, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "test",
		AuthCode:   "pass",
		AuthType:   constants.AuthTypePassword,
		VerifyType: "INVALID",
	})
	if err == nil {
		t.Fatal("expected error for invalid verify type")
	}
}

func TestLogin_InvalidAuthType(t *testing.T) {
	reset()
	_, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "test",
		AuthCode:   "pass",
		AuthType:   "INVALID",
		VerifyType: constants.VerifyTypeUsername,
	})
	if err == nil || !errors.Is(err, lumerrors.ErrInvalidAuthType) {
		t.Fatalf("expected ErrInvalidAuthType error, got: %v", err)
	}
}

// ========= Login - VERIFY_CODE 测试 =========

func TestLoginVerifyCode_ExistingUser_Success(t *testing.T) {
	reset()
	addTestUser("user4", "user4@example.com", "", "", 1)
	mockCommonService.VerifyCode = "123456"

	resp, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "user4@example.com",
		AuthCode:   "123456",
		AuthType:   constants.AuthTypeVerifyCode,
		VerifyType: constants.VerifyTypeEmail,
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.UserID == 0 {
		t.Fatal("expected non-zero UserID")
	}
	if resp.Token == "" {
		t.Fatal("expected JWT token for verify-code login")
	}
}

func TestLoginVerifyCode_WrongCode(t *testing.T) {
	reset()
	addTestUser("user5", "user5@example.com", "", "", 1)
	mockCommonService.VerifyCode = "123456"

	_, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "user5@example.com",
		AuthCode:   "000000",
		AuthType:   constants.AuthTypeVerifyCode,
		VerifyType: constants.VerifyTypeEmail,
	})
	if err == nil {
		t.Fatal("expected error for wrong verify code")
	}
}

func TestLoginVerifyCode_UserNotFound_NoAutoCreate(t *testing.T) {
	reset()
	mockCommonService.VerifyCode = "654321"

	_, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier:        "newuser@example.com",
		AuthCode:          "654321",
		AuthType:          constants.AuthTypeVerifyCode,
		VerifyType:        constants.VerifyTypeEmail,
		DeviceFingerprint: "device-1",
		UserAgent:         "test-agent",
		IP:                "127.0.0.1",
	})
	if err == nil || !errors.Is(err, lumerrors.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound (no auto-create), got: %v", err)
	}
	// Verify user was NOT created
	if _, err := mockUserDao.FindByEmail(context.Background(), "newuser@example.com"); err == nil {
		t.Fatal("expected no user to be auto-created on verify-code login")
	}
}

// ========= Login - THIRD_PARTY 测试 =========

func TestLoginThirdParty_MatchByOpenID(t *testing.T) {
	reset()
	user := addTestUser("tpuser1", "", "", "", 1)
	mockThirdPartyDao.Create(context.Background(), &model.ThirdPartyUserMeta{
		UserID:   user.ID,
		Provider: "github",
		OpenID:   "github-open-123",
	})
	RegisterMockThirdPartyAdapter("github", NewMockThirdPartyAdapter(&model.ThirdPartyUser{
		Provider:    "github",
		OpenID:      "github-open-123",
		AccessToken: "github-token",
	}))

	resp, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier:  "github-open-123",
		AuthCode:    "github-open-123",
		AuthType:    constants.AuthTypeThirdParty,
		VerifyType:  "github",
		AccessToken: "github-token",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.UserID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, resp.UserID)
	}
	if resp.Token == "" {
		t.Fatal("expected JWT token for third-party login")
	}
}

func TestLoginThirdParty_MatchByUnionID(t *testing.T) {
	reset()
	user := addTestUser("tpuser2", "", "", "", 1)
	mockThirdPartyDao.Create(context.Background(), &model.ThirdPartyUserMeta{
		UserID:   user.ID,
		Provider: "wechat",
		OpenID:   "wx-open-old",
		UnionID:  "wx-union-456",
	})
	RegisterMockThirdPartyAdapter("wechat", NewMockThirdPartyAdapter(&model.ThirdPartyUser{
		Provider:    "wechat",
		OpenID:      "wx-open-new", // different from existing
		UnionID:     "wx-union-456", // matches existing by union
		AccessToken: "wx-token",
	}))

	resp, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier:  "wx-open-new",
		AuthCode:    "code",
		AuthType:    constants.AuthTypeThirdParty,
		VerifyType:  "wechat",
		AccessToken: "wx-token",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.UserID != user.ID {
		t.Fatalf("expected user ID %d (matched by union), got %d", user.ID, resp.UserID)
	}
}

func TestLoginThirdParty_MatchByEmail(t *testing.T) {
	reset()
	user := addTestUser("tpuser3", "tp3@example.com", "", "", 1)
	RegisterMockThirdPartyAdapter("github", NewMockThirdPartyAdapter(&model.ThirdPartyUser{
		Provider:    "github",
		OpenID:      "github-open-999",
		Email:       "tp3@example.com", // matches by email
		AccessToken: "token",
	}))

	resp, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier:  "github-open-999",
		AuthCode:    "code",
		AuthType:    constants.AuthTypeThirdParty,
		VerifyType:  "github",
		AccessToken: "token",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.UserID != user.ID {
		t.Fatalf("expected user ID %d (matched by email), got %d", user.ID, resp.UserID)
	}
}

func TestLoginThirdParty_AutoCreateUser(t *testing.T) {
	reset()
	RegisterMockThirdPartyAdapter("github", NewMockThirdPartyAdapter(&model.ThirdPartyUser{
		Provider:    "github",
		OpenID:      "github-new-openid",
		AccessToken: "new-token",
	}))

	resp, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier:        "github-new-openid",
		AuthCode:          "code",
		AuthType:          constants.AuthTypeThirdParty,
		VerifyType:        "github",
		AccessToken:       "new-token",
		DeviceFingerprint: "device-tp",
		UserAgent:         "agent",
		IP:                "10.0.0.1",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.UserID == 0 {
		t.Fatal("expected non-zero UserID for auto-created user")
	}
}

// ========= Register 测试 =========

func TestRegister_Success(t *testing.T) {
	reset()
	mockCommonService.VerifyCode = "888888"

	resp, err := getUserService().Register(context.Background(), &v1.RegisterRequest{
		Identifier:        "newuser",
		AuthCode:          "888888",
		Password:          "mypassword",
		AuthType:          constants.AuthTypeVerifyCode,
		VerifyType:        constants.VerifyTypeUsername,
		DeviceFingerprint: "device-new",
		UserAgent:         "agent",
		IP:                "10.0.0.1",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if resp.UserID == 0 {
		t.Fatal("expected non-zero UserID")
	}
	if resp.Token == "" {
		t.Fatal("expected JWT token in register response")
	}
}

func TestRegister_DuplicateUser(t *testing.T) {
	reset()
	mockCommonService.VerifyCode = "999999"
	addTestUser("existing", "", "", "", 1)

	_, err := getUserService().Register(context.Background(), &v1.RegisterRequest{
		Identifier: "existing",
		AuthCode:   "999999",
		Password:   "pass123",
		AuthType:   constants.AuthTypeVerifyCode,
		VerifyType: constants.VerifyTypeUsername,
	})
	if err == nil || !errors.Is(err, lumerrors.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists error, got: %v", err)
	}
}

func TestRegister_InvalidVerifyType(t *testing.T) {
	reset()
	mockCommonService.VerifyCode = "123456"
	_, err := getUserService().Register(context.Background(), &v1.RegisterRequest{
		Identifier: "test",
		AuthCode:   "123456",
		Password:   "pass",
		AuthType:   constants.AuthTypeVerifyCode,
		VerifyType: "INVALID",
	})
	if err == nil {
		t.Fatal("expected error for invalid verify type")
	}
}

// ========= GetUserInfo 测试 =========

func TestGetUserInfo_Success(t *testing.T) {
	reset()
	user := addTestUser("infouser", "info@example.com", "13900139000", "pass", 1)

	result, err := getUserService().GetUserInfo(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetUserInfo failed: %v", err)
	}
	if result.Username != "infouser" {
		t.Fatalf("expected username infouser, got %s", result.Username)
	}
	if result.Email != "info@example.com" {
		t.Fatalf("expected email info@example.com, got %s", result.Email)
	}
}

func TestGetUserInfo_NotFound(t *testing.T) {
	reset()
	_, err := getUserService().GetUserInfo(context.Background(), 99999)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

// ========= UpdateUserInfo 测试 =========

func TestUpdateUserInfo_Success(t *testing.T) {
	reset()
	user := addTestUser("updateuser", "old@example.com", "", "pass", 1)

	result, err := getUserService().UpdateUserInfo(context.Background(), user.ID, "新昵称", "18800188000", "http://avatar.url")
	if err != nil {
		t.Fatalf("UpdateUserInfo failed: %v", err)
	}
	if result.Nickname != "新昵称" {
		t.Fatalf("expected nickname '新昵称', got %s", result.Nickname)
	}
	if result.Phone != "18800188000" {
		t.Fatalf("expected phone '18800188000', got %s", result.Phone)
	}
	if result.Avatar != "http://avatar.url" {
		t.Fatalf("expected avatar 'http://avatar.url', got %s", result.Avatar)
	}
}

func TestUpdateUserInfo_Partial(t *testing.T) {
	reset()
	user := addTestUser("partial", "partial@example.com", "13900139000", "pass", 1)

	result, err := getUserService().UpdateUserInfo(context.Background(), user.ID, "onlynick", "", "")
	if err != nil {
		t.Fatalf("UpdateUserInfo failed: %v", err)
	}
	if result.Nickname != "onlynick" {
		t.Fatalf("expected nickname 'onlynick', got %s", result.Nickname)
	}
	if result.Phone != "13900139000" {
		t.Fatalf("phone should remain '13900139000', got %s", result.Phone)
	}
}

func TestUpdateUserInfo_NotFound(t *testing.T) {
	reset()
	_, err := getUserService().UpdateUserInfo(context.Background(), 99999, "nick", "", "")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

// ========= ChangePassword 测试 =========

func TestChangePassword_Success(t *testing.T) {
	reset()
	user := addTestUser("changepwd", "cp@example.com", "", "oldPass1", 1)

	err := getUserService().ChangePassword(context.Background(), user.ID, "oldPass1", "newPass1")
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Verify new password works
	_, err = getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "changepwd",
		AuthCode:   "newPass1",
		AuthType:   constants.AuthTypePassword,
		VerifyType: constants.VerifyTypeUsername,
	})
	if err != nil {
		t.Fatal("expected login with new password to succeed")
	}
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	reset()
	user := addTestUser("cpuser", "", "", "correctOld", 1)

	err := getUserService().ChangePassword(context.Background(), user.ID, "wrongOld", "newPass")
	if err == nil || !errors.Is(err, lumerrors.ErrOldPasswordWrong) {
		t.Fatalf("expected ErrOldPasswordWrong error, got: %v", err)
	}
}

func TestChangePassword_UserNotFound(t *testing.T) {
	reset()
	err := getUserService().ChangePassword(context.Background(), 99999, "old", "new")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

// ========= ResetPassword 测试 =========

func TestResetPassword_Success(t *testing.T) {
	reset()
	addTestUser("resetuser", "reset@example.com", "", "oldPwd", 1)
	mockCommonService.VerifyCode = "112233"

	err := getUserService().ResetPassword(context.Background(), "reset@example.com", "112233", "newPwd123")
	if err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	// Verify new password works
	_, err = getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier: "reset@example.com",
		AuthCode:   "newPwd123",
		AuthType:   constants.AuthTypePassword,
		VerifyType: constants.VerifyTypeEmail,
	})
	if err != nil {
		t.Fatal("expected login with new password to succeed")
	}
}

func TestResetPassword_WrongVerifyCode(t *testing.T) {
	reset()
	addTestUser("rpuser", "rp@example.com", "", "old", 1)
	mockCommonService.VerifyCode = "111111"

	err := getUserService().ResetPassword(context.Background(), "rp@example.com", "000000", "newPass")
	if err == nil {
		t.Fatal("expected error for wrong verify code")
	}
}

func TestResetPassword_UserNotFound(t *testing.T) {
	reset()
	mockCommonService.VerifyCode = "123456"

	err := getUserService().ResetPassword(context.Background(), "noone@example.com", "123456", "newPass")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

// ========= SendResetPasswordCode 测试 =========

func TestSendResetPasswordCode_Success(t *testing.T) {
	reset()
	addTestUser("sendcode", "sendcode@example.com", "", "pass", 1)

	err := getUserService().SendResetPasswordCode(context.Background(), "sendcode@example.com")
	if err != nil {
		t.Fatalf("SendResetPasswordCode failed: %v", err)
	}
}

func TestSendResetPasswordCode_UserNotFound(t *testing.T) {
	reset()
	err := getUserService().SendResetPasswordCode(context.Background(), "noone@example.com")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

// ========= LoginRecord 测试 =========

func TestLoginRecord_NewDevice(t *testing.T) {
	reset()
	user := addTestUser("recorduser", "", "", "pass", 1)

	getUserService().LoginRecord(context.Background(), &model.LoginRecordRequest{
		UserID:            user.ID,
		IP:                "192.168.1.1",
		DeviceFingerprint: "device-record-1",
		UserAgent:         "Mozilla/5.0",
		Status:            1,
		Device:            "iPhone 15",
	})

	// Verify login history was recorded
	histories, _, err := mockLoginHistoryDao.FindByUserID(context.Background(), user.ID, 1, 10)
	if err != nil {
		t.Fatalf("expected login history, got error: %v", err)
	}
	if len(histories) != 1 {
		t.Fatalf("expected 1 login history, got %d", len(histories))
	}
	if histories[0].IP != "192.168.1.1" {
		t.Fatalf("expected IP 192.168.1.1, got %s", histories[0].IP)
	}

	// Verify trusted device was created
	device, err := mockTrustedDeviceDao.FindByUserIDAndFingerprint(context.Background(), user.ID, "device-record-1")
	if err != nil {
		t.Fatalf("expected trusted device, got error: %v", err)
	}
	if device.DeviceName != "iPhone 15" {
		t.Fatalf("expected device name 'iPhone 15', got %s", device.DeviceName)
	}
}

func TestLoginRecord_ExistingDevice(t *testing.T) {
	reset()
	user := addTestUser("recorduser2", "", "", "pass", 1)

	// First call - create device
	getUserService().LoginRecord(context.Background(), &model.LoginRecordRequest{
		UserID:            user.ID,
		IP:                "10.0.0.1",
		DeviceFingerprint: "device-existing",
		UserAgent:         "curl",
		Status:            1,
		Device:            "Chrome",
	})

	// Second call - update device (different IP)
	getUserService().LoginRecord(context.Background(), &model.LoginRecordRequest{
		UserID:            user.ID,
		IP:                "10.0.0.2",
		DeviceFingerprint: "device-existing",
		UserAgent:         "curl",
		Status:            1,
		Device:            "Chrome",
	})

	// Verify device was updated
	device, err := mockTrustedDeviceDao.FindByUserIDAndFingerprint(context.Background(), user.ID, "device-existing")
	if err != nil {
		t.Fatalf("expected trusted device, got error: %v", err)
	}
	if device.LastLoginIP != "10.0.0.2" {
		t.Fatalf("expected updated IP 10.0.0.2, got %s", device.LastLoginIP)
	}
}

// ========= 综合场景测试 =========

func TestLogin_NewDeviceAndDifferentLocation(t *testing.T) {
	reset()
	addTestUser("secureuser", "secure@example.com", "", "pass123", 1)

	resp, err := getUserService().Login(context.Background(), &v1.LoginRequest{
		Identifier:        "secureuser",
		AuthCode:          "pass123",
		AuthType:          constants.AuthTypePassword,
		VerifyType:        constants.VerifyTypeUsername,
		DeviceFingerprint: "new-device",
		IP:                "10.0.0.5",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.UserID == 0 {
		t.Fatal("expected non-zero UserID")
	}
	// DifferentLocation is always true in current impl
	if !resp.DifferentLocation {
		t.Fatal("expected DifferentLocation=true (always returns true)")
	}
}
