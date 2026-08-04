# Changelog

本项目所有重要变更均记录在此文件中。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

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

## [Unreleased - v0.5.0 适配]

### Changed
- Go 工具链升级至 `go 1.26` + `toolchain go1.26.5`（landc-go v0.5.0 要求）
- landc-go 依赖版本统一至 `v0.5.0`（replace 指令在 v0.5.0 正式发布后移除）
- GitHub Actions 各阶段 go-version 升级至 `1.26`
- 依赖随 landc-go 一并升级，修复已知 CVE
