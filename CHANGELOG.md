# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- ✨ 签到配置管理页面 (`AdminCheckInConfigsPage.jsx`)
- ✨ 结构化日志模块 (`internal/logger`)
- ✨ 数据库连接池配置 (`DBConfig`)
- ✨ Redis 连接验证 (`NewRedisWithPing`)

### Changed
- 🔧 使用结构化日志替代 `fmt.Println` 输出
- 🔧 所有管理页面添加 Toast 错误/成功提示
- 🔧 所有删除操作添加 Popconfirm 二次确认
- 🔧 数据库连接现在使用连接池优化
- 🔧 Redis 连接启动时验证连通性
- 🔧 所有管理页面表格添加分页支持
- 🔧 用户个人中心页面添加 Toast 提示

### Fixed
- 🐛 修复 `admin_routes.go` 中多处错误处理缺少 `return` 语句的问题
- 🐛 删除 `server.go` 中重复的注释代码块

## [1.0.0] - 2025-11-24

### Added
- 🎉 首次正式发布
- ✨ LinuxDo OAuth 登录集成
- ✨ 双重认证机制（JWT Token 和 API Key）
- ✨ 配额限流系统（基于用户等级和模型前缀）
- ✨ 积分系统（按模型计费，支持预扣和退款）
- ✨ 每日签到功能（连续签到统计，余额衰减机制）
- ✨ 渠道管理（多上游渠道，模型唯一性约束）
- ✨ 完整日志系统（API 调用、登录、操作、积分交易）
- ✨ 管理后台（用户、渠道、配额规则、积分规则、日志、统计）
- 🐳 Docker 和 Docker Compose 支持
- 📝 完整的文档（README、ADMIN_GUIDE、CONTRIBUTING）
- 🔄 GitHub Actions CI/CD（自动构建、测试、发布）
- 🧪 单元测试覆盖（后端和前端）

### Technical Details
- 后端：Go 1.23 + Gin + GORM + PostgreSQL + Redis
- 前端：React 18 + Vite + Semi UI
- 数据库迁移：4 个 SQL 脚本
- 多平台支持：Linux (amd64/arm64)、macOS (amd64/arm64)、Windows (amd64)

---

## 版本说明格式

### Added
- 新功能

### Changed
- 功能变更

### Deprecated
- 即将废弃的功能

### Removed
- 已移除的功能

### Fixed
- Bug 修复

### Security
- 安全相关更新

[Unreleased]: https://github.com/dbccccccc/linuxdo-relay/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/dbccccccc/linuxdo-relay/releases/tag/v1.0.0
