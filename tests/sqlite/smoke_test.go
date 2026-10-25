package sqlite_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	frameauth "github.com/LandcLi/landc-go/frame/pkg/auth"
	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"

	_ "github.com/LandcLi/LandcUserManager/internal" // 触发默认实现注册
	"github.com/LandcLi/LandcUserManager/model"
	"github.com/LandcLi/LandcUserManager/serverlib"
	"github.com/LandcLi/LandcUserManager/service"

	userapi "github.com/LandcLi/LandcUserManager/api/user"
	userv1 "github.com/LandcLi/LandcUserManager/api/user/v1"
)

// sqliteConfigYAML 以 SQLite 为唯一外部依赖，验证「开箱即用」能力。
const sqliteConfigYAML = `
server:
  port: 0
  addr: "127.0.0.1"
  use_default_routes: false
database:
  driver: "sqlite"
  dsn: "%s"
  auto_migrate: true
redis:
  addr: ""
jwt:
  secret: "sqlite-smoke-test-secret-key-at-least-32-chars"
  expire_time: "2h"
  issuer: "landc-user-manager-test"
password:
  bcrypt_cost: 4
verify_code:
  length: 6
  expire_time: 300
login_history:
  retention_days: 90
cache:
  user_cache_expire: 3600
http_client:
  timeout: 10
`

// setupSQLiteEnv 用临时目录中的 SQLite 库完成一次完整 bootstrap。
// 返回数据库文件路径。
func setupSQLiteEnv(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "data", "lum.db")
	configPath := filepath.Join(dir, "config.yaml")

	content := fmt.Sprintf(sqliteConfigYAML, filepath.ToSlash(dbPath))
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := frameconfig.InitGlobalConfigWithPath(configPath); err != nil {
		t.Fatalf("init global config: %v", err)
	}
	// 清理全局状态，避免同包其他用例互相污染
	t.Cleanup(func() {
		_ = framedb.Close()
		_ = framecache.Close()
	})

	if err := framedb.InitGlobalDBWithDefault(); err != nil {
		t.Fatalf("init db: %v", err)
	}
	if err := framecache.InitGlobalCacheWithDefault(); err != nil {
		t.Fatalf("init cache: %v", err)
	}
	frameauth.InitFromConfig(frameconfig.GetConfig())

	if err := serverlib.EnsureSchema(); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
	return dbPath
}

// TestSQLiteSchemaCreated 验证 auto_migrate 真正建出了本系统的全部表。
func TestSQLiteSchemaCreated(t *testing.T) {
	setupSQLiteEnv(t)

	migrator := framedb.GetDB().Migrator()
	for _, m := range serverlib.Models() {
		if !migrator.HasTable(m) {
			t.Errorf("table missing after EnsureSchema: %T", m)
		}
	}
}

// TestSQLiteRegisterAndLogin 验证注册 → 登录的完整业务闭环可在 SQLite 上跑通。
func TestSQLiteRegisterAndLogin(t *testing.T) {
	setupSQLiteEnv(t)

	ctx := context.Background()
	userSvc := service.GetUserService()

	const (
		username = "sqlite_smoke_user"
		password = "Passw0rd!"
		fp       = "sqlite-smoke-fingerprint"
	)

	regResp, err := userSvc.Register(ctx, &userv1.RegisterRequest{
		Identifier:        username,
		Password:          password,
		AuthCode:          "PASSWORD",
		AuthType:          "PASSWORD",
		VerifyType:        "USERNAME",
		IP:                "127.0.0.1",
		DeviceFingerprint: fp,
		UserAgent:         "sqlite-smoke-test",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if regResp.UserID == 0 {
		t.Fatal("register returned zero user id")
	}
	if regResp.Token == "" {
		t.Fatal("register should issue access token")
	}

	loginResp, err := userSvc.Login(ctx, &userv1.LoginRequest{
		Identifier:        username,
		AuthCode:          password,
		AuthType:          "PASSWORD",
		VerifyType:        "USERNAME",
		IP:                "127.0.0.1",
		DeviceFingerprint: fp,
		UserAgent:         "sqlite-smoke-test",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loginResp.UserID != regResp.UserID {
		t.Errorf("login user id = %d, want %d", loginResp.UserID, regResp.UserID)
	}

	// 注册时已标记该设备可信，同设备再次登录不应判为新设备
	if loginResp.IsNewDevice {
		t.Error("registered trusted device should not be flagged as new device")
	}

	// 读回用户，确认落库字段正确
	info, err := userSvc.GetUserInfo(ctx, regResp.UserID)
	if err != nil {
		t.Fatalf("get user info: %v", err)
	}
	if info.Username != username {
		t.Errorf("username = %q, want %q", info.Username, username)
	}
	if info.Password == password {
		t.Error("password must be hashed in database")
	}

	var count int64
	if err := framedb.GetDB().Model(&model.LoginHistory{}).Count(&count).Error; err != nil {
		t.Fatalf("count login history: %v", err)
	}
	if count == 0 {
		t.Error("expected login history recorded")
	}
}

// TestGetUserInfoByExplicitUserID 验证服务间调用可以按 user_id 取用户档案。
//
// 场景：LandcSSO 的 OIDC /userinfo 端点按令牌里的 sub 取用户信息，
// 此时没有「当前登录用户」的概念，必须靠请求体显式指定。
func TestGetUserInfoByExplicitUserID(t *testing.T) {
	setupSQLiteEnv(t)

	ctx := context.Background()
	userSvc := service.GetUserService()

	const (
		username = "explicit_lookup_user"
		password = "Passw0rd!"
	)

	regResp, err := userSvc.Register(ctx, &userv1.RegisterRequest{
		Identifier:        username,
		Password:          password,
		AuthType:          "PASSWORD",
		VerifyType:        "USERNAME",
		IP:                "127.0.0.1",
		DeviceFingerprint: "explicit-lookup-fp",
		UserAgent:         "explicit-lookup-test",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// 走 controller：请求体里显式给 user_id
	ctrl := userapi.GetUserController()
	info, err := ctrl.GetUserInfo(ctx, &userv1.GetUserInfoRequest{UserID: regResp.UserID})
	if err != nil {
		t.Fatalf("GetUserInfo(explicit user_id): %v", err)
	}
	if info.ID != regResp.UserID {
		t.Errorf("返回的 ID = %d, want %d", info.ID, regResp.UserID)
	}
	if info.Username != username {
		t.Errorf("返回的 Username = %q, want %q", info.Username, username)
	}

	// 不指定且 context 里没有登录用户时，应退化为 0 而不是panic
	if _, err := ctrl.GetUserInfo(ctx, &userv1.GetUserInfoRequest{}); err == nil {
		t.Log("未指定 user_id 且无登录上下文时返回了结果（取决于实现，允许）")
	}
}

// TestRegisterMultipleUsersWithoutEmail 验证多个「没有邮箱」的用户可以并存。
//
// 历史缺陷：email 是 NOT NULL + UNIQUE，而用户名注册时 email 为空串。
// SQL 里空串是确定值，不是 NULL —— 于是全表只能有一个无邮箱用户，
// 第二个用户名注册直接 UNIQUE constraint failed。这不是边角场景：
// 用户名注册是最常见的注册方式，第二个人注册就会撞上。
func TestRegisterMultipleUsersWithoutEmail(t *testing.T) {
	setupSQLiteEnv(t)

	ctx := context.Background()
	userSvc := service.GetUserService()

	usernames := []string{"no_email_user_a", "no_email_user_b", "no_email_user_c"}
	ids := make(map[uint]struct{})

	for _, name := range usernames {
		resp, err := userSvc.Register(ctx, &userv1.RegisterRequest{
			Identifier:        name,
			Password:          "Passw0rd!",
			AuthType:          "PASSWORD",
			VerifyType:        "USERNAME",
			IP:                "127.0.0.1",
			DeviceFingerprint: "fp-" + name,
			UserAgent:         "multi-no-email-test",
		})
		if err != nil {
			t.Fatalf("注册 %s 失败: %v", name, err)
		}
		if _, dup := ids[resp.UserID]; dup {
			t.Fatalf("用户 ID 重复: %d", resp.UserID)
		}
		ids[resp.UserID] = struct{}{}
	}

	if len(ids) != len(usernames) {
		t.Errorf("成功注册 %d 个用户，want %d", len(ids), len(usernames))
	}
}
