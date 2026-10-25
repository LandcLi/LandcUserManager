# Changelog

本项目所有重要变更均记录在此文件中。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [0.1.3] - 2026-10-25

### 变更

- 依赖升级到 `landc-go/frame v0.8.6`（含 `api v0.5.1`、`tools v0.7.1`）。
  该版本修掉了「`other:` 段业务配置从未生效」与「HTTP 状态码被日志中间件吞成 200」
  两个框架级缺陷 —— 后者会让参数校验失败、鉴权失败在网关与 APM 里全都显示为成功。
- CI 拆成两个 job：`CGO_ENABLED=0` 跑测试（对齐发布产物形态），
  `CGO_ENABLED=1` 单独跑 `-race`。原先用 cgo 环境跑完测试、却产出静态无 cgo 的二进制，
  两边约束不一致。

本版本无代码变更，仅依赖与流水线调整。

## [0.1.2] - 2026-10-25

> 说明：本段同时收录了 v0.1.1 发布时未单独记录的安全修复内容。

### 变更

- 依赖升级到 `landc-go/frame v0.8.5`。该版本修复了两个会让服务
  「看起来在跑、实际不可用」的框架缺陷：配置文件省略字段退化为零值
  （导致 `health_check.enabled`、`use_default_routes` 被静默关闭），
  以及 JWT 鉴权中间件拦截健康检查端点。升级后 `/health`、`/ready`
  无需 Token 即可访问，编排系统探针可用。

### 修复（安全）

- **新设备判定语义写反**：`Login` 中 `isNewDevice` 直接取自 `detectTrustedDevice`
  （该方法返回「设备可信」），导致**可信设备被判为新设备、真正的陌生设备反而不触发**。
  由于 SSO 依赖该字段决定是否下发临时令牌，实际后果是「新设备二次验证」防护完全失效。
  现改为 `req.DeviceFingerprint != "" && !detectTrustedDevice(...)`。
- **异地登录检测硬编码 `return true`**：每次登录都被判定为异地，风控信号形同虚设。
  现与最近一次登录的 IP 比较，首次登录（无历史）不判为异地。

### Added

- **SQLite 支持（默认开箱即用）**：`database.driver: sqlite` + `auto_migrate: true`
  即可零外部依赖启动，首次运行自动建库建表。纯 Go 驱动，无需 CGO/gcc。
- **`serverlib.EnsureSchema()` / `serverlib.Models()`**：本系统的模型清单与建表入口
  集中在一处；`RegisterToRouter` 会先确保表结构存在（受 `database.auto_migrate` 控制），
  库模式下宿主无需额外调用。
- **`LoginHistoryDao.FindLatestByUserID`**：供异地登录判定使用。
- **`tests/sqlite` 集成冒烟**：临时目录真实 SQLite，覆盖建表、注册、登录、
  密码哈希落库、登录历史记录与设备信任判定。
- `config.example.yml` 默认改为 SQLite，并附 MySQL/Redis 切换说明。

### Changed

- 依赖升级至 `landc-go/frame v0.8.4`（SQLite 纯 Go 驱动、`auto_migrate`、库模式路由注册）。
- README 重写：SQLite 快速开始、库模式嵌入、API 与响应约定、配置要点。

### Added
- 新增 `pkg/errors` 包：预定义类型化错误变量，支持 `errors.Is()/errors.As()` 判断
- 新增 `pkg/consts` 包：统一管理认证类型、验证类型、缓存 Key 等魔法字符串
- 新增 `utility/crypto` 包：AES-GCM 敏感字段加密（用于第三方 AccessToken）
- 登录/注册成功后签发 JWT（access + refresh token），完成认证闭环
- 全局鉴权中间件：公开路径白名单放行，其余路径强制校验 Bearer Token
- 登录成功后自动记录登录历史与信任设备
- 验证码发送限流：发送间隔 + 每日次数上限
- 新增 Dockerfile 多阶段构建、Makefile、.golangci.yml、GitHub Actions CI
- 新增 CONTRIBUTING.md / CODE_OF_CONDUCT.md / SECURITY.md / CHANGELOG.md

### Changed
- 硬编码中文错误字符串重构为 `pkg/errors` 错误变量
- 修复验证码 Action 大小写不一致（发送/校验 Key 不匹配）的核心 Bug
- 验证码登录不再自动注册用户（防止验证码爆破批量开户）
- 获取/更新用户信息、修改密码、登录记录的 UserID 改从 JWT claims 解析（修复 IDOR 越权）
- 启动时初始化第三方登录适配器（`InitWithConfig`），第三方登录可用
- 修复自定义配置解析（`UnmarshalOther` + json tag），ThirdParty/Email/SMS 等配置可正确加载
- 修复配置拼写错误 `ThridParty` → `ThirdParty`
- 第三方登录 AccessToken 加密存储
- config.example.yml 中 JWT/邮件字段与框架标准对齐
- Makefile 去除本机绝对路径，支持环境变量覆盖
- 工具链锁定 go1.24.9

### Deprecated
- 无

### Removed
- 无

### Fixed
- 见 Changed 中的 Bug 修复项

### Security
- 登录/注册/敏感接口不再信任前端传入的 UserID（防 IDOR）
- 第三方 AccessToken 不再明文入库
- JWT 密钥强度校验（>= 32 字符）

## [Unreleased - v0.8.1 适配]

### Changed
- landc-go `frame` 升级至 **v0.8.1**（修复 LUM 提交的「web 业务错误码透传」需求）
- `pkg/errors` 错误变量统一改用 `core.NewError`（带业务错误码）：用户不存在 → `40400`、密码错误 → `40100`、验证码错误 → `42201`、限流 → `50300` 等
- `middleware.Auth` 将 claims 注入 request context（`createHandler` 传给 controller 的是 `c.Request.Context()`），controller 从标准 ctx 读取 UserID

## [Unreleased - v0.7.0 适配]

### Changed
- landc-go `tools`/`frame` 升级至 **v0.7.0**（对应提取需求 A1/B1/B2 落地）
- **A1 加解密**：`utility/crypto` 重构为 `tools/security` 薄封装（密钥来源 `LUM_ENC_KEY` 由 LUM 侧封装，底层 AES-GCM 来自 tools/security）
- **B1 验证码**：`SendVerifyCode`/`CheckVerifyCode` 切换为 `frame/pkg/verifycode`（生成/存储/TTL/发送间隔/每日上限/一次性校验由框架承担），移除 `common_impl` 手写限流与验证码逻辑
- **B2 限流**：验证码限流底层使用 `tools/ratelimit`（缓存原子 Incr，并发不丢计数）

### Removed
- `common_impl` 手写限流逻辑（`getDailyCount`/`incrementDailyCount`/间隔标记）——由框架能力替代

## [Unreleased - v0.6.0 适配]

### Added
- `serverlib` 库模式入口：`RegisterToRouter`（`WithAuth`/`WithWebOptions`），独立部署与下游嵌入共用同一代码路径
- landc-go v0.6.0 能力验证测试：注册选项（V1）/ 显式 proxy 构造（V3）/ OpenAPI（V4）/ 命名资源作用域（V5）/ 模拟宿主嵌入（L3）

### Changed
- DAO/service 层全量迁移为 ctx 感知（`migrate-dbctx`：`GetDB()`→`GetDBFrom(ctx)`、`GetCache()`→`GetCacheFrom(ctx)`）
- `landc gen lib` 生成 `serverlib` 骨架并补全 common 控制器 + 鉴权中间件 + `InitWithConfig`
- `landc gen proxy` 重新生成 SDK（显式 `NewUserControllerProxy`/`NewCommonControllerProxy` 构造函数）
- landc-go `frame` 升级至 **v0.6.0**（注册选项/路由查询/命名资源作用域/库模式入口），api/log/tools 保持 v0.5.0，**移除开发期 replace**
- 独立部署入口 `internal/cmd` 改用 `serverlib.RegisterToRouter`

## [Unreleased - v0.5.0 适配]

### Changed
- Go 工具链升级至 `go 1.26` + `toolchain go1.26.5`（landc-go v0.5.0 要求）
- landc-go 依赖统一至线上 `v0.5.0`，**移除本地 replace**（landc-go 已于 2026-08-04 公开发布）
- 后续 landc-go 更新走正常版本发布，通过 `go get -u` 升级
- GitHub Actions 各阶段 go-version 升级至 `1.26`，移除本地 landc-go clone 步骤
- Dockerfile 升级至 `golang:1.26-alpine`，移除本地 landc-go 复制步骤
- 依赖随 landc-go 一并升级，修复已知 CVE
