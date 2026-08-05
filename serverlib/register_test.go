package serverlib

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	"github.com/LandcLi/landc-go/frame/pkg/web"
	"github.com/gin-gonic/gin"
)

var bootstrapOnce sync.Once

func bootstrap(t *testing.T) {
	t.Helper()
	bootstrapOnce.Do(func() {
		if err := frameconfig.InitGlobalConfigWithDefault(); err != nil {
			t.Fatalf("init global config: %v", err)
		}
	})
}

func findRoute(s *web.Server, handlerName string) (web.RouteInfo, bool) {
	for _, r := range s.Routes() {
		if r.HandlerName == handlerName {
			return r, true
		}
	}
	return web.RouteInfo{}, false
}

func TestRegisterToRouterRegistersRoutes(t *testing.T) {
	bootstrap(t)
	gin.SetMode(gin.TestMode)
	s := web.New()

	if err := RegisterToRouter(s); err != nil {
		t.Fatalf("RegisterToRouter: %v", err)
	}

	for _, name := range []string{"Login", "Register", "GetUserInfo"} {
		if _, ok := findRoute(s, name); !ok {
			t.Errorf("routes missing handler %s", name)
		}
	}
}

func TestAppendRoutesAfterRegister(t *testing.T) {
	// 只装配、不接管：注册后仍可继续追加自有路由。
	// 注意：默认 WithAuth(true) 时全局 Auth 会影响宿主后续追加的路由（见设计文档 4.2/风险#9），
	// 此测试用 WithAuth(false) 聚焦"可继续追加路由"的核心语义。
	bootstrap(t)
	gin.SetMode(gin.TestMode)
	s := web.New()

	if err := RegisterToRouter(s, WithAuth(false)); err != nil {
		t.Fatalf("RegisterToRouter: %v", err)
	}

	// 追加自有路由：既可用 web 的 Group，也可用 gin 原生 engine（此处演示后者）
	s.Engine().Group("/sso").GET("/callback", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sso/callback", http.NoBody)
	s.Engine().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /sso/callback = %d, want 200", w.Code)
	}
}

func TestLoginRouteReachable(t *testing.T) {
	// 关闭鉴权后，Login 路由应可达（handler 执行返回 400，而非 404）
	bootstrap(t)
	gin.SetMode(gin.TestMode)
	s := web.New()

	if err := RegisterToRouter(s, WithAuth(false)); err != nil {
		t.Fatalf("RegisterToRouter: %v", err)
	}

	login, ok := findRoute(s, "Login")
	if !ok {
		t.Fatal("routes missing Login")
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(login.Method, login.Path, http.NoBody)
	s.Engine().ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Errorf("Login route %s %s not reachable (404)", login.Method, login.Path)
	}
}
