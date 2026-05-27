# LandcUserManager

一个功能完善的用户管理系统，提供用户注册、登录、信息管理、第三方登录等功能，支持多种认证方式和灵活的配置管理。

## 功能特性

### 用户管理
- 用户注册（支持邮箱、手机号、用户名）
- 用户登录（密码登录、验证码登录、第三方登录）
- 用户信息查询和更新
- 密码修改和重置
- 登录记录管理
- 设备指纹识别和信任设备管理

### 第三方登录支持
- 微信登录
- GitHub登录
- 支付宝登录
- 钉钉登录

### 安全特性
- 设备指纹识别
- 信任设备管理
- 异地登录检测
- 验证码限流
- 密码强度验证（bcrypt加密）

### 通用功能
- 邮箱验证码发送
- 手机验证码发送
- 短信验证码发送
- 登录历史记录

## 技术栈

- **框架**: landc-go/mvc（DI / Gateway / Web）
- **Web框架**: Gin
- **数据库**: MySQL (GORM)
- **缓存**: Redis
- **日志**: landc-go/log
- **API规范**: landc-go/api
- **工具集**: landc-go/tools
- **配置管理**: YAML
- **加密**: bcrypt, crypto

## 项目结构

```
LandcUserManager/
├── api/                    # API接口定义
│   ├── common/             # 通用接口（验证码等）
│   └── user/               # 用户相关接口
├── dao/                    # 数据访问层接口
├── internal/                # 内部实现
│   ├── controller/         # 控制器层
│   ├── dao_impl/          # DAO实现层
│   ├── middleware/        # 中间件
│   └── service_impl/      # 服务实现层
├── model/                  # 数据模型
├── sdk/                    # HTTP SDK
├── service/                # 服务层接口
├── sqls/                   # SQL脚本
├── utility/                # 工具包
│   ├── client/            # HTTP客户端
│   ├── config/           # 配置管理
│   ├── db/               # 数据库连接
│   └── logger/           # 日志管理
└── LandcUserManager.go     # 主入口文件
```

## 快速开始

### 1. 配置文件

复制示例配置文件并修改：

```bash
cp config.example.yml config.yml
```

配置文件包含以下部分：

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 60
  write_timeout: 60
  gin_mode: "debug"
  enable_cors: true
  user_default_route: true

database:
  dsn: "user:password@tcp(localhost:3306)/landc_user_manager?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
  max_retries: 3
  dial_timeout: 5
  read_timeout: 5
  write_timeout: 5

jwt:
  secret_key: "your-secret-key"
  private_key_path: ""
  public_key_path: ""
  expires_in: 7200
  refresh_in: 604800

email:
  host: "smtp.example.com"
  port: 587
  username: "your-email@example.com"
  password: "your-password"
  from: "your-email@example.com"

sms:
  api_key: "your-api-key"
  api_secret: "your-api-secret"
  sign_name: "your-sign-name"
  template: "your-template-code"

logger:
  level: "info"
  format: "json"
  output: "stdout"

cache:
  user_cache_expire: 3600

password:
  bcrypt_cost: 10

verify_code:
  length: 6
  expire_time: 300

http_client:
  timeout: 30

login_history:
  retention_days: 90

rate_limit:
  verify_code_interval: 60
  max_verify_code_per_day: 10

third_party:
  wechat:
    app_id: "your-wechat-app-id"
    app_secret: "your-wechat-app-secret"
  github:
    client_id: "your-github-client-id"
    client_secret: "your-github-client-secret"
  alipay:
    app_id: "your-alipay-app-id"
    app_secret: "your-alipay-app-secret"
  dingtalk:
    app_key: "your-dingtalk-app-key"
    app_secret: "your-dingtalk-app-secret"
```

### 2. 初始化数据库

执行SQL脚本初始化数据库表结构：

```bash
mysql -u root -p < sqls/init.sql
```

### 3. 启动服务

```go
package main

import (
	"github.com/LandcLi/LandcUserManager"
)

func main() {
	// 使用配置文件初始化
	engine, err := landcusermanager.InitLUMDefault("config.yml")
	if err != nil {
		panic(err)
	}

	// 启动服务
	engine.Run(":8080")
}
```

## 使用SDK（远程调用模式）

当 LandcUserManager 独立部署时，其他服务通过 SDK 远程访问：

```go
package main

import (
	"context"
	"fmt"
	"github.com/LandcLi/LandcUserManager/api/user"
	"github.com/LandcLi/LandcUserManager/api/user/v1"
	"github.com/LandcLi/LandcUserManager/sdk"
)

func main() {
	// 创建远程 Controller（基于 landc-go/mvc Gateway 模式）
	remote := sdk.NewRemoteUserController("http://localhost:8080")

	// 注册为远程实现（替代本地实现）
	user.UserGateway.Override(remote)

	// 使用方式与本地完全一致
	ctrl := user.GetUserController()
	ctx := context.Background()
	resp, err := ctrl.Login(ctx, &v1.LoginRequest{
		Identifier:        "user@example.com",
		AuthCode:          "123456",
		VerifyType:        "EMAIL",
		AuthType:          "VERIFY_CODE",
		IP:                "127.0.0.1",
		DeviceFingerprint: "device123",
		UserAgent:         "Mozilla/5.0",
	})
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}
	fmt.Printf("登录成功，用户ID: %d\n", resp.UserID)
}
```

### 本地/远程切换

本项目基于 `landc-go/mvc` 的 **Gateway 模式**，只需在 `main.go` 中修改注册方式即可切换：

```go
// 本地模式（默认，通过 init() 自动注册）
landcusermanager.InitLUMWithConfigAndGin(cfg, ginEngine)

// 远程模式（替换为 SDK 远程调用）
user.UserGateway.Override(sdk.NewRemoteUserController("http://user-service:8081"))
```

业务代码（如 SSO 的 auth controller）**完全不需要修改**。

## API接口

### 用户接口

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 登录 | POST | /api/user/login | 用户登录 |
| 注册 | POST | /api/user/register | 用户注册 |
| 获取用户信息 | POST | /api/user/get-user-info | 获取用户信息 |
| 更新用户信息 | POST | /api/user/update-user-info | 更新用户信息 |
| 修改密码 | POST | /api/user/change-password | 修改密码 |
| 重置密码 | POST | /api/user/reset-password | 重置密码 |
| 登录记录 | POST | /api/user/login-record | 记录登录信息 |

### 通用接口

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 发送验证码 | POST | /api/common/send-verify-code | 发送验证码 |

## 初始化方法

### InitLUMDefault

从配置文件路径初始化，自动加载配置并初始化所有组件。

```go
engine, err := landcusermanager.InitLUMDefault("config.yml")
```

### InitLUMWithConfig

使用配置对象初始化，支持自定义配置。

```go
cfg := &config.Config{
    // 自定义配置
}
engine, err := landcusermanager.InitLUMWithConfig(cfg)
```

### InitLUMWithConfigAndGin

使用配置对象和自定义Gin引擎初始化，支持更灵活的集成。

```go
cfg := &config.Config{
    // 自定义配置
}
engine := gin.Default()
err := landcusermanager.InitLUMWithConfigAndGin(cfg, engine)
```

## 第三方登录

支持多种第三方登录方式，通过适配器模式实现，方便扩展。

### 微信登录

```go
wechatAdapter := third_party.NewWechatAdapter(&config.ThirdPartyGlobalConfig{
    ThirdParty: map[string]map[string]interface{}{
        "wechat": {
            "app_id":     "your-app-id",
            "app_secret": "your-app-secret",
        },
    },
})
```

### GitHub登录

```go
githubAdapter := third_party.NewGithubAdapter(&config.ThirdPartyGlobalConfig{
    ThirdParty: map[string]map[string]interface{}{
        "github": {
            "client_id":     "your-client-id",
            "client_secret": "your-client-secret",
        },
    },
})
```

## 安全特性

### 设备指纹识别
- 自动识别用户设备
- 支持信任设备管理
- 新设备登录提醒

### 验码限流
- 防止验证码滥用
- 支持发送间隔控制
- 每日发送次数限制

### 密码安全
- bcrypt加密存储
- 密码强度验证
- 支持密码修改和重置

## 贡献

欢迎提交Issue和Pull Request来帮助改进这个项目。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 联系方式

如有问题或建议，欢迎通过以下方式联系：

- 提交 Issue
- 发送邮件

## 更新日志

### v1.0.0
- 初始版本发布
- 支持用户注册、登录、信息管理
- 支持多种第三方登录
- 提供完整的HTTP SDK
- 支持灵活的配置管理
