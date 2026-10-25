package sqlite_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LandcLi/landc-go/api/core"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"

	userv1 "github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/model"
	lumerrors "github.com/LandcLi/LandcUserManager/pkg/errors"
	"github.com/LandcLi/LandcUserManager/service"
)

// errorCode 取出框架标准错误里的业务码；非标准错误返回 0。
func errorCode(t *testing.T, err error) core.ErrorCode {
	t.Helper()
	if err == nil {
		return core.CodeSuccess
	}
	var apiErr *core.Error
	if !errors.As(err, &apiErr) {
		return 0
	}
	return apiErr.Code
}

// errorMessage 取出框架标准错误里的文案。
func errorMessage(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		return ""
	}
	var apiErr *core.Error
	if errors.As(err, &apiErr) {
		return apiErr.Message
	}
	return err.Error()
}

// loginWithPassword 用密码登录，返回错误（成功时为 nil）。
func loginWithPassword(ctx context.Context, identifier, password, verifyType string) error {
	_, err := service.GetUserService().Login(ctx, &userv1.LoginRequest{
		Identifier:        identifier,
		AuthCode:          password,
		AuthType:          "PASSWORD",
		VerifyType:        verifyType,
		IP:                "127.0.0.1",
		DeviceFingerprint: "lockout-test-fp",
		UserAgent:         "lockout-test",
	})
	return err
}

// registerUser 注册一个用户并返回 userID。
func registerUser(t *testing.T, username, password string) uint {
	t.Helper()
	resp, err := service.GetUserService().Register(context.Background(), &userv1.RegisterRequest{
		Identifier:        username,
		Password:          password,
		AuthType:          "PASSWORD",
		VerifyType:        "USERNAME",
		IP:                "127.0.0.1",
		DeviceFingerprint: "lockout-test-fp",
		UserAgent:         "lockout-test",
	})
	if err != nil {
		t.Fatalf("注册 %s 失败: %v", username, err)
	}
	return resp.UserID
}

// TestPasswordLoginLocksAfterRepeatedFailures 验证连续密码错误会触发临时锁定。
//
// 历史行为：密码错了只是返回「密码错误」，可以无限次重试 —— 6 位数字口令
// 几小时就能被在线爆破。现在第 max_failed_attempts 次失败即锁定，
// 且锁定期内**正确密码也进不去**（否则锁定形同虚设）。
func TestPasswordLoginLocksAfterRepeatedFailures(t *testing.T) {
	setupSQLiteEnv(t)

	ctx := context.Background()
	const (
		username = "lockout_user"
		password = "Passw0rd!"
	)
	userID := registerUser(t, username, password)

	// 默认策略：5 次失败锁定。前 4 次是普通凭据错误。
	for i := 1; i < 5; i++ {
		err := loginWithPassword(ctx, username, "wrong-password", "USERNAME")
		if code := errorCode(t, err); code != core.ErrorCodeUnauthorized {
			t.Fatalf("第 %d 次错误密码返回 code=%d, want %d（err=%v）",
				i, code, core.ErrorCodeUnauthorized, err)
		}
	}

	// 第 5 次失败即触发锁定。
	err := loginWithPassword(ctx, username, "wrong-password", "USERNAME")
	if code := errorCode(t, err); code != lumerrors.ErrorCodeAccountLocked {
		t.Fatalf("触发锁定的那次失败返回 code=%d, want %d（err=%v）",
			code, lumerrors.ErrorCodeAccountLocked, err)
	}

	// 锁定期内用正确密码也必须被拒（这是锁定的意义所在）。
	err = loginWithPassword(ctx, username, password, "USERNAME")
	if code := errorCode(t, err); code != lumerrors.ErrorCodeAccountLocked {
		t.Fatalf("锁定期内正确密码返回 code=%d, want %d（err=%v）",
			code, lumerrors.ErrorCodeAccountLocked, err)
	}

	// 失败历史应留痕：5 条 status=0 的记录，且带标识符。
	var histories []model.LoginHistory
	if err := framedb.GetDB().Where("identifier = ? AND status = 0", username).
		Find(&histories).Error; err != nil {
		t.Fatalf("查询失败登录历史: %v", err)
	}
	if len(histories) != 5 {
		t.Errorf("失败登录记录 %d 条, want 5", len(histories))
	}
	if len(histories) > 0 {
		if histories[0].UserID != userID {
			t.Errorf("失败记录 user_id = %d, want %d", histories[0].UserID, userID)
		}
		if histories[0].IP == "" {
			t.Error("失败记录应保留 IP 便于取证")
		}
	}

	// 锁定期内被直接拒绝的请求不应再落库，否则可被用来放大攻击撑爆历史表。
	var total int64
	if err := framedb.GetDB().Model(&model.LoginHistory{}).
		Where("identifier = ?", username).Count(&total).Error; err != nil {
		t.Fatalf("统计登录历史: %v", err)
	}
	// 5 次失败 + 2 次锁定期内被拒（不该落库）= 期望仍是 5 条
	if total != 5 {
		t.Errorf("登录历史共 %d 条, want 5（锁定期内的请求不应落库）", total)
	}
}

// TestLoginDoesNotRevealAccountExistence 验证「账号不存在」与「密码错误」
// 返回完全相同的错误，堵住用户名枚举。
//
// 历史行为：不存在 → code 40400 "user not found"，密码错 → code 40100
// "password incorrect"。两个码、两段文案，攻击者拿一份用户名清单跑一遍
// 就能把「哪些账号真实存在」筛出来，为后续撞库定向准备。
func TestLoginDoesNotRevealAccountExistence(t *testing.T) {
	setupSQLiteEnv(t)

	ctx := context.Background()
	const (
		username = "enum_probe_user"
		password = "Passw0rd!"
	)
	registerUser(t, username, password)

	notFoundErr := loginWithPassword(ctx, "definitely_not_exist_9c1f", "whatever", "USERNAME")
	wrongPwdErr := loginWithPassword(ctx, username, "wrong-password", "USERNAME")

	if notFoundErr == nil || wrongPwdErr == nil {
		t.Fatalf("两种错误场景都应失败: notFound=%v wrongPwd=%v", notFoundErr, wrongPwdErr)
	}
	if errorCode(t, notFoundErr) != errorCode(t, wrongPwdErr) {
		t.Errorf("错误码泄露账号是否存在: 不存在=%d 密码错=%d",
			errorCode(t, notFoundErr), errorCode(t, wrongPwdErr))
	}
	if errorMessage(t, notFoundErr) != errorMessage(t, wrongPwdErr) {
		t.Errorf("错误文案泄露账号是否存在: 不存在=%q 密码错=%q",
			errorMessage(t, notFoundErr), errorMessage(t, wrongPwdErr))
	}
	if errorCode(t, wrongPwdErr) != core.ErrorCodeUnauthorized {
		t.Errorf("凭据错误应为 40100, got %d", errorCode(t, wrongPwdErr))
	}
}

// TestSuccessfulLoginResetsFailureCount 验证成功登录清零计数。
//
// 若不清零，「整个生命周期累计 5 次失败」就会锁人 —— 正常用户打错几次键盘
// 迟早被莫名其妙锁掉，这类策略没人敢开。
func TestSuccessfulLoginResetsFailureCount(t *testing.T) {
	setupSQLiteEnv(t)

	ctx := context.Background()
	const (
		username = "reset_counter_user"
		password = "Passw0rd!"
	)
	registerUser(t, username, password)

	// 4 次失败（未达阈值）
	for i := 0; i < 4; i++ {
		if err := loginWithPassword(ctx, username, "wrong-password", "USERNAME"); errorCode(t, err) != core.ErrorCodeUnauthorized {
			t.Fatalf("第 %d 次失败返回 %v", i+1, err)
		}
	}
	// 成功一次 → 计数清零
	if err := loginWithPassword(ctx, username, password, "USERNAME"); err != nil {
		t.Fatalf("正确密码登录失败: %v", err)
	}
	// 再失败 4 次仍不应被锁
	for i := 0; i < 4; i++ {
		err := loginWithPassword(ctx, username, "wrong-password", "USERNAME")
		if errorCode(t, err) != core.ErrorCodeUnauthorized {
			t.Fatalf("清零后再失败第 %d 次被锁: %v", i+1, err)
		}
	}
	// 依然可以登录
	if err := loginWithPassword(ctx, username, password, "USERNAME"); err != nil {
		t.Fatalf("计数清零后正确密码应可登录: %v", err)
	}
}

// TestIdentifyCaseIsNormalizedForLockout 验证大小写不敏感。
//
// 计数按标识符归一化保存：如果 "Alice" 与 "alice" 各算一份，
// 攻击者只要换个大小写就能把 5 次重试上限翻倍。
func TestIdentifyCaseIsNormalizedForLockout(t *testing.T) {
	setupSQLiteEnv(t)

	ctx := context.Background()
	const (
		username = "case_probe_user"
		password = "Passw0rd!"
	)
	registerUser(t, username, password)

	// 用不同大小写各失败 1 次，累计 4 次（若未归一化，会分成 4 份各 1 次）
	variants := []string{username, "CASE_PROBE_USER", "Case_Probe_User", "case_probe_user"}
	for i, v := range variants {
		err := loginWithPassword(ctx, v, "wrong-password", "USERNAME")
		if code := errorCode(t, err); code != core.ErrorCodeUnauthorized {
			t.Fatalf("第 %d 次（%s）失败返回 code=%d, want %d（err=%v）",
				i+1, v, code, core.ErrorCodeUnauthorized, err)
		}
	}

	// 第 5 次（换个大小写）应触发锁定 —— 说明计数确实按归一化后的标识符累计
	err := loginWithPassword(ctx, "CASE_probe_user", "wrong-password", "USERNAME")
	if code := errorCode(t, err); code != lumerrors.ErrorCodeAccountLocked {
		t.Fatalf("大小写变体未共享失败计数，code=%d, want %d（err=%v）",
			code, lumerrors.ErrorCodeAccountLocked, err)
	}
}

// TestChangePasswordClearsLockout 验证改密后解除临时锁定。
//
// 用户被撞库打得锁定之后，本人改完密码还要干等 15 分钟才能登录，
// 这在真实支持场景里很难解释。改密成功说明已证明身份，应当立即解封。
func TestChangePasswordClearsLockout(t *testing.T) {
	setupSQLiteEnv(t)

	ctx := context.Background()
	const (
		username = "unlock_after_change"
		oldPwd   = "Passw0rd!"
		newPwd   = "NewPassw0rd!"
	)
	userID := registerUser(t, username, oldPwd)

	// 打到锁定
	for i := 0; i < 5; i++ {
		_ = loginWithPassword(ctx, username, "wrong-password", "USERNAME")
	}
	if code := errorCode(t, loginWithPassword(ctx, username, oldPwd, "USERNAME")); code != lumerrors.ErrorCodeAccountLocked {
		t.Fatalf("前置条件不成立：账号未被锁定（code=%d）", code)
	}

	if err := service.GetUserService().ChangePassword(ctx, userID, oldPwd, newPwd); err != nil {
		t.Fatalf("改密失败: %v", err)
	}

	if err := loginWithPassword(ctx, username, newPwd, "USERNAME"); err != nil {
		t.Fatalf("改密后应解除锁定并可直接登录: %v", err)
	}
}

// TestSendResetPasswordCodeHidesUnknownEmail 验证找回密码不暴露邮箱是否注册。
//
// 未注册的邮箱必须返回成功（客户端统一提示「如果该邮箱已注册，验证码已发送」），
// 否则「忘记密码」就成了一个无需登录的邮箱枚举接口。
func TestSendResetPasswordCodeHidesUnknownEmail(t *testing.T) {
	setupSQLiteEnv(t)

	ctx := context.Background()
	if err := service.GetUserService().SendResetPasswordCode(ctx, "nobody-here@example.com"); err != nil {
		t.Fatalf("未注册邮箱不应报错（防枚举），got: %v", err)
	}
}
