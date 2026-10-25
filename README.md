# LandcUserManager

统一用户中心（LUM）：用户注册、认证、信息管理与第三方登录，是 LandcSSO 的底层用户数据源。

## 功能

| 领域 | 能力 |
| --- | --- |
| 账号 | 注册（用户名/邮箱/手机号）、密码登录、验证码登录、第三方登录 |
| 资料 | 查询、更新、改密、重置密码 |
| 安全 | bcrypt 密码哈希、设备指纹、可信设备、新设备与异地登录识别、验证码限流 |
| 记录 | 登录历史、可信设备列表 |
| 第三方 | 微信、GitHub、支付宝、钉钉 |
| 通用 | 邮箱/短信验证码发送 |

## 技术栈

- **框架**：`github.com/LandcLi/landc-go/frame`（配置 / DB / 缓存 / JWT / DI / Web / 健康检查）
- **Web**：Gin；路由由 `meta.Meta` 标签声明，框架自动注册
- **数据库**：SQLite（默认，纯 Go 驱动，零外部依赖）或 MySQL（生产）
- **缓存**：Redis（可选，缺失时自动降级为进程内 LRU）
- **日志**：`landc-go/log`

## 快速开始

```bash
cp config.example.yml config.yaml
go run .
```

默认使用 SQLite，首次启动会在 `data/landc_user_manager.db` 自动建库建表，无需 MySQL/Redis。
启动后：

```bash
curl http://127.0.0.1:18081/health
```

注册一个用户（`authType: PASSWORD` 时无需验证码）：

```bash
curl -X POST http://127.0.0.1:18081/api/user/register \
  -H 'Content-Type: application/json' \
  -d '{
    "identifier": "admin",
    "password": "Passw0rd!",
    "verifyCode": "PASSWORD",
    "authType": "PASSWORD",
    "verifyType": "USERNAME",
    "ip": "127.0.0.1",
    "deviceFingerprint": "local-dev",
    "userAgent": "curl"
  }'
```

### 切换到 MySQL / Redis（生产）

```yaml
database:
  driver: "mysql"
  dsn: "user:password@tcp(127.0.0.1:3306)/landc_user_manager?charset=utf8mb4&parseTime=True&loc=Local"
  auto_migrate: false          # 生产由 sqls/init.sql 建表

redis:
  addr: "localhost:6379"
```

```bash
mysql -u root -p < sqls/init.sql
```

### 作为库嵌入（进程内模式）

LandcSSO 通过库模式复用本系统，无需额外 HTTP 跳转：

```go
import (
    "github.com/LandcLi/LandcUserManager/serverlib"
    "github.com/LandcLi/landc-go/frame/pkg/web"
)

server := web.New()
if err := serverlib.RegisterToRouter(server); err != nil {  // 含建表、鉴权中间件、第三方适配器
    return err
}
```

宿主进程需先完成框架 bootstrap（配置 / DB / 缓存 / JWT）。`serverlib` 只装配、不接管：
不会创建 engine、不会调用 `Run`，调用方仍可继续追加自有路由。

## API

所有接口为 `POST` + JSON。

| 路径 | 说明 | 鉴权 |
| --- | --- | --- |
| `/api/user/register` | 注册并直接签发令牌 | 否 |
| `/api/user/login` | 登录（密码 / 验证码 / 第三方） | 否 |
| `/api/common/send-verify-code` | 发送邮箱或短信验证码 | 否 |
| `/api/user/reset-password` | 验证码重置密码 | 否 |
| `/api/user/get-user-info` | 查询当前用户 | 是 |
| `/api/user/update-user-info` | 更新资料 | 是 |
| `/api/user/change-password` | 修改密码 | 是 |
| `/api/user/login-record` | 记录登录历史与可信设备 | 是 |

响应约定：成功直接返回业务对象（HTTP 200），不额外包 `{code,data}` 外壳；
失败返回非 2xx + `{"code":<业务错误码>,"message":"..."}`（框架统一错误体）。

`Login` 响应中的 `isNewDevice` / `differentLocation` 供 SSO 决定是否下发临时令牌：
新设备或异地登录需二次兑换，详见 `LandcSSO` 的临时令牌机制。

## 配置

配置项见 `config.example.yml`，要点：

- `database.auto_migrate: true` 时启动自动建表（开发/测试），生产建议关闭；
- `redis.addr` 留空即使用进程内缓存，仅单实例部署可用；
- `jwt.secret` 长度需 ≥ 32，生产用环境变量 `LANDC_JWT_SECRET` 注入；
- `email` / `sms` 未配置时对应验证码渠道不可用（密码登录不受影响）。

## 目录结构

```
api/                 接口定义（请求/响应 + meta 路由标签）
dao/                 数据访问接口
internal/
  controller/        控制器（注册到 DI）
  dao_impl/          DAO 实现
  service_impl/      业务实现
  middleware/        鉴权中间件
  cmd/               独立部署入口
model/               数据模型
serverlib/           库模式入口（含 EnsureSchema / Models）
service/             服务接口
sdk/                 HTTP 远程 SDK
sqls/                MySQL 初始化脚本
tests/               单元测试（mock）与 tests/sqlite（SQLite 集成冒烟）
```

## 测试

```bash
go test ./...            # 单元测试 + SQLite 集成冒烟
go vet ./...
```

`tests/sqlite` 使用临时目录中的真实 SQLite 文件，覆盖建表、注册、登录、密码哈希与登录历史落库。

## 许可证

MIT，见 [LICENSE](LICENSE)。
