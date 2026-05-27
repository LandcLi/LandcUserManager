package landcusermanager

import (
	"github.com/LandcLi/LandcUserManager/api"
	_ "github.com/LandcLi/LandcUserManager/internal"
	"github.com/LandcLi/LandcUserManager/utility/client"
	"github.com/LandcLi/LandcUserManager/utility/config"
	"github.com/LandcLi/LandcUserManager/utility/db"
	"github.com/LandcLi/LandcUserManager/utility/logger"
	apiMiddleware "github.com/LandcLi/landc-go/api/middleware/gin"
	ginLogger "github.com/LandcLi/landc-go/log/adapter/gin"
	"github.com/gin-gonic/gin"
)

// InitLUMDefault 从配置文件路径初始化
func InitLUMDefault(configPath string) (*gin.Engine, error) {
	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// 初始化Gin引擎
	ginEngine := gin.Default()

	// 配置Gin
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	// 启用CORS
	if cfg.Server.EnableCORS {
		ginEngine.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		})
	}

	// 初始化服务
	// 依次调用初始化方法
	logger.InitLogger(&cfg.Log)
	_, _ = db.InitRedis(&cfg.Redis)
	db.InitDatabase(&cfg.Database)
	client.InitMailClientByConfig(cfg.Email.Config)

	ginEngine.Use(apiMiddleware.GinMiddleware())
	ginLogger.UseWithGin(ginEngine, logger.GetLogger())

	// 注册默认路由
	if cfg.Server.UserDefaultRoute {
		api.RegisterDefaultRoutes(ginEngine)
	}

	return ginEngine, nil
}

// InitLUMWithConfig 从配置对象初始化
func InitLUMWithConfig(cfg *config.Config) (*gin.Engine, error) {
	// 初始化Gin引擎
	ginEngine := gin.Default()

	// 配置Gin
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	// 启用CORS
	if cfg.Server.EnableCORS {
		ginEngine.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		})
	}

	// 初始化服务
	// 依次调用初始化方法
	logger.InitLogger(&cfg.Log)
	_, _ = db.InitRedis(&cfg.Redis)
	db.InitDatabase(&cfg.Database)
	client.InitMailClientByConfig(cfg.Email.Config)

	ginEngine.Use(apiMiddleware.GinMiddleware())
	ginLogger.UseWithGin(ginEngine, logger.GetLogger())

	// 注册默认路由
	if cfg.Server.UserDefaultRoute {
		api.RegisterDefaultRoutes(ginEngine)
	}

	return ginEngine, nil
}

// InitLUMWithConfigAndGin 使用指定的配置和Gin引擎初始化
func InitLUMWithConfigAndGin(cfg *config.Config, ginEngine *gin.Engine) error {
	// 配置Gin
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	// 启用CORS
	if cfg.Server.EnableCORS {
		ginEngine.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		})
	}

	// 初始化服务
	// 依次调用初始化方法
	logger.InitLogger(&cfg.Log)
	_, _ = db.InitRedis(&cfg.Redis)
	db.InitDatabase(&cfg.Database)
	client.InitMailClientByConfig(cfg.Email.Config)

	ginEngine.Use(apiMiddleware.GinMiddleware())
	ginLogger.UseWithGin(ginEngine, logger.GetLogger())

	// 注册默认路由
	if cfg.Server.UserDefaultRoute {
		api.RegisterDefaultRoutes(ginEngine)
	}

	return nil
}
