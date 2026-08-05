// 需求单验证测试：landc-go v0.6.0 能力逐项实测（V1/V3/V4/V5，V2/V6 已由既有 serverlib 测试覆盖）。
package serverlib

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/frame/pkg/meta"
	"github.com/LandcLi/landc-go/frame/pkg/openapi"
	"github.com/LandcLi/landc-go/frame/pkg/resource"
	"github.com/LandcLi/landc-go/frame/pkg/web"
	"github.com/gin-gonic/gin"

	userApi "github.com/LandcLi/LandcUserManager/api/user"
	"github.com/LandcLi/LandcUserManager/sdk"
)

// ---- 最小验证控制器（不依赖 LUM 业务接口，聚焦框架能力） ----

type helloRequest struct {
	meta.Meta `path:"/api/hello" method:"POST"`
	Name      string `json:"name"`
}

type helloResponse struct {
	DBName string `json:"dbName"`
	Cache  bool   `json:"cache"`
	Config string `json:"config"`
}

type helloController struct{}

// Hello 在 handler 内通过 GetXxxFrom(ctx) 解析资源，验证作用域注入。
func (c *helloController) Hello(ctx context.Context, req *helloRequest) (*helloResponse, error) {
	dbName := "global"
	cfgName := "global"
	cacheHit := false

	if s, ok := resource.FromContext(ctx); ok {
		dbName = s.DB
		cfgName = s.Config
	}

	// 命名 DB 生效验证：解析到的连接不是全局连接（隔离）
	gdb := framedb.GetDBFrom(ctx)
	if gdb == nil {
		return nil, fmt.Errorf("GetDBFrom(ctx) is nil")
	}
	if gdb != framedb.GetDB() {
		dbName = dbName + "|isolated"
	}

	if framecache.GetCacheFrom(ctx) != nil {
		cacheHit = true
	}

	return &helloResponse{DBName: dbName, Cache: cacheHit, Config: cfgName}, nil
}

// ---- bootstrap ----

var verifyOnce sync.Once

func verifyBootstrap(t *testing.T) {
	t.Helper()
	verifyOnce.Do(func() {
		if err := frameconfig.InitGlobalConfigWithDefault(); err != nil {
			t.Fatalf("init global config: %v", err)
		}
		// 注册命名资源：sqlite 内存库 + 本地缓存
		if err := framedb.InitNamedDB("lum-db", &frameconfig.DatabaseConfig{
			Driver: "sqlite",
			DSN:    "file::memory:?cache=shared",
		}); err != nil {
			t.Fatalf("init named db: %v", err)
		}
		framecache.InitNamedCacheWithLocal("lum-cache", 100)
	})
}

// ---- V1：注册选项（前缀 / 单方法覆盖 / 多中间件） ----

func TestVerifyV1_RegisterOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := web.New()

	var mwOrder []string
	mw1 := func(c *gin.Context) { mwOrder = append(mwOrder, "mw1"); c.Next() }
	mw2 := func(c *gin.Context) { mwOrder = append(mwOrder, "mw2"); c.Next() }
	ctrlMW := func(c *gin.Context) { mwOrder = append(mwOrder, "ctrl"); c.Next() }

	err := s.RegisterHandler(&helloController{},
		web.WithPrefix("/api/v2"),
		web.WithMethodPath("Hello", "/custom/hello"),
		web.WithMethodHTTPMethod("Hello", "PUT"),
		web.WithMethodMiddleware("Hello", mw1, mw2),
		web.WithControllerMiddleware(ctrlMW),
	)
	if err != nil {
		t.Fatalf("RegisterHandler: %v", err)
	}

	found := false
	for _, r := range s.Routes() {
		if r.HandlerName == "Hello" {
			t.Logf("route: %s %s", r.Method, r.Path)
			if r.Method == "PUT" && r.Path == "/api/v2/custom/hello" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected PUT /api/v2/custom/hello after options (prefix + method path override)")
	}

	// 中间件顺序：控制器级先于方法级，方法级按声明顺序
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v2/custom/hello", http.NoBody)
	s.Engine().ServeHTTP(w, req)
	want := []string{"ctrl", "mw1", "mw2"}
	if strings.Join(mwOrder, ",") != strings.Join(want, ",") {
		t.Fatalf("middleware order = %v, want %v", mwOrder, want)
	}
}

// ---- V3：显式 proxy 构造函数 ----

func TestVerifyV3_ExplicitProxyConstructor(t *testing.T) {
	up := sdk.NewUserControllerProxy("http://localhost:18081")
	if up == nil {
		t.Fatal("NewUserControllerProxy returned nil")
	}
	cp := sdk.NewCommonControllerProxy("http://localhost:18081")
	if cp == nil {
		t.Fatal("NewCommonControllerProxy returned nil")
	}
}

// ---- V4：OpenAPI 自动收集与落盘 ----

func TestVerifyV4_OpenAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := web.New()
	if err := s.RegisterHandler(&helloController{}); err != nil {
		t.Fatalf("RegisterHandler: %v", err)
	}

	g := openapi.NewGenerator(openapi.Info{Title: "LUM verify", Version: "v0.6.0-verify"})
	g.RegisterServer(s)
	data, err := g.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if !strings.Contains(string(data), "/api/hello") {
		t.Fatal("openapi JSON missing /api/hello path")
	}

	// WriteFile：CI 落盘 openapi.json
	tmp := filepath.Join(t.TempDir(), "openapi.json")
	if err := g.WriteFile(tmp); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := os.Stat(tmp); err != nil {
		t.Fatalf("WriteFile output missing: %v", err)
	}
}

// ---- V5：命名资源作用域 ----

func TestVerifyV5_NamedScopeInjection(t *testing.T) {
	verifyBootstrap(t)
	gin.SetMode(gin.TestMode)
	s := web.New()

	if err := s.RegisterLibrary(&helloController{}, web.WithScope(resource.Scope{
		Name:  "lum",
		DB:    "lum-db",
		Cache: "lum-cache",
	})); err != nil {
		t.Fatalf("RegisterLibrary: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/hello", http.NoBody)
	s.Engine().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "lum-db|isolated") {
		t.Fatalf("expected named db isolation, body = %s", body)
	}
	if !strings.Contains(body, `"cache":true`) {
		t.Fatalf("expected named cache resolution, body = %s", body)
	}
}

// ---- L3：模拟宿主嵌入（真实 LUM 控制器 + 命名 DB + 追加自有路由） ----

func TestVerifyL3_HostEmbedding(t *testing.T) {
	verifyBootstrap(t)
	gin.SetMode(gin.TestMode)
	s := web.New()

	// 模拟宿主：用 RegisterLibrary 注册真实 LUM UserController，指定命名 DB（LUM 独立库）
	if err := s.RegisterLibrary(userApi.GetUserController(), web.WithScope(resource.Scope{
		Name: "lum",
		DB:   "lum-db",
	})); err != nil {
		t.Fatalf("RegisterLibrary: %v", err)
	}

	// 宿主追加自有路由（只装配、不接管）
	s.Engine().Group("/sso").GET("/callback", func(c *gin.Context) {
		c.String(http.StatusOK, "sso-callback-ok")
	})

	// LUM 路由可达（scope 注入后不 404；NoBody 触发 binding 失败返回 400 属预期）
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/user/login", http.NoBody)
	s.Engine().ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Fatal("LUM login route not reachable after RegisterLibrary")
	}

	// 宿主自有路由可达
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/sso/callback", http.NoBody)
	s.Engine().ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("host route /sso/callback = %d, want 200", w2.Code)
	}
}

func TestVerifyV5_UnregisteredScopeError(t *testing.T) {
	verifyBootstrap(t)
	gin.SetMode(gin.TestMode)
	s := web.New()

	err := s.RegisterLibrary(&helloController{}, web.WithScope(resource.Scope{
		Name: "lum",
		DB:   "no-such-db",
	}))
	if err == nil {
		t.Fatal("expected error for unregistered named db (no silent fallback)")
	}
	if !strings.Contains(err.Error(), "no-such-db") {
		t.Fatalf("unexpected error: %v", err)
	}
}
