package landcusermanager

import (
	"fmt"

	"github.com/LandcLi/LandcUserManager/api"
	_ "github.com/LandcLi/LandcUserManager/internal"
	"github.com/LandcLi/LandcUserManager/service"
	"github.com/LandcLi/LandcUserManager/utility/client"
	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/LandcLi/LandcUserManager/utility/logger"
	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/frame/pkg/middleware"
	"github.com/LandcLi/landc-go/frame/pkg/web"
	"github.com/gin-gonic/gin"
)

// InitLUMDefault 从配置文件路径初始化，返回 web.Server。
// 框架全权管理配置加载，LUM 仅通过 config.GetConfig() 做类型化访问。
func InitLUMDefault(configPath string) (*web.Server, error) {
	if err := frameconfig.InitGlobalConfigWithPathAndEnv(configPath); err != nil {
		return nil, fmt.Errorf("框架加载配置失败: %w", err)
	}
	return initLUM(config.GetConfig())
}

// InitLUMWithConfig 从配置对象初始化，返回 web.Server
func InitLUMWithConfig(cfg *config.Config) (*web.Server, error) {
	return initLUM(cfg)
}

// InitLUMWithConfigAndGin 使用指定配置和已有 Gin 引擎初始化（兼容模式）
func InitLUMWithConfigAndGin(cfg *config.Config, ginEngine *gin.Engine) error {
	return initLUMOnGin(cfg, ginEngine)
}

func initLUM(cfg *config.Config) (*web.Server, error) {
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	server := web.NewServer(&web.ServerConfig{
		Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
	})

	if err := initInfrastructure(cfg); err != nil {
		return nil, err
	}

	if cfg.Server.UserDefaultRoute {
		registerRoutes(server)
	}

	return server, nil
}

func initLUMOnGin(cfg *config.Config, ginEngine *gin.Engine) error {
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	if err := initInfrastructure(cfg); err != nil {
		return err
	}

	if cfg.Server.UserDefaultRoute {
		registerRoutesOnGin(ginEngine)
	}

	return nil
}

func initInfrastructure(cfg *config.Config) error {
	logger.InitLogger(&cfg.Log)

	framedb.InitGlobalDBWithDefault()
	framecache.InitGlobalCacheWithDefault()

	client.InitMailClientByConfig(cfg.Email.Config)

	if err := service.InitWithConfig(&cfg.ThridParty); err != nil {
		return fmt.Errorf("初始化第三方登录适配器失败: %w", err)
	}

	return nil
}

func registerRoutes(server *web.Server) {
	engine := server.Engine()
	engine.Use(middleware.Trace(), middleware.CORS(), middleware.Recovery())
	api.RegisterDefaultRoutes(engine)
}

func registerRoutesOnGin(ginEngine *gin.Engine) {
	ginEngine.Use(middleware.Trace(), middleware.CORS(), middleware.Recovery())
	api.RegisterDefaultRoutes(ginEngine)
}
