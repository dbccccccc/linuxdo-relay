# linuxdo-relay

一个独立的 API 转发服务，基于 Go + Gin + PostgreSQL + Redis

前端控制台使用 React + Vite + Semi UI，提供个人中心和管理员功能。

## 功能特性

- ✅ **LinuxDo OAuth 登录**：集成 LinuxDo 账号体系
- ✅ **双重认证**：支持 JWT Token（Web）和 API Key（程序调用）
- ✅ **配额限流**：基于用户等级和模型前缀的灵活限流策略
- ✅ **积分系统**：按模型计费，支持预扣和失败退款
- ✅ **每日签到**：积分奖励，连续签到统计
- ✅ **渠道管理**：多上游渠道，模型唯一性约束
- ✅ **完整日志**：API 调用、登录、操作记录
- ✅ **管理后台**：用户管理、渠道配置、规则设置、数据统计

## 快速开始

### Docker Compose（推荐）

1. **克隆项目并配置**
```bash
git clone https://github.com/dbccccccc/linuxdo-relay.git
cd linuxdo-relay
cp .env.example .env
# 编辑 .env 配置 OAuth 和 JWT 密钥
```

2. **启动服务**
```bash
docker-compose up -d
```

程序会自动连接数据库并执行表结构迁移，无需手动操作。

3. **访问服务**：http://localhost:8080

### 使用 Docker 镜像

```bash
docker run -d -p 8080:8080 \
  -e APP_PG_DSN="postgres://user:pass@host:5432/db?sslmode=disable" \
  -e APP_REDIS_ADDR="redis:6379" \
  -e APP_JWT_SECRET="your-secret-at-least-32-chars" \
  -e APP_LINUXDO_CLIENT_ID="your-client-id" \
  -e APP_LINUXDO_CLIENT_SECRET="your-client-secret" \
  -e APP_LINUXDO_REDIRECT_URL="http://localhost:8080/auth/linuxdo/callback" \
  ghcr.io/dbccccccc/linuxdo-relay:latest
```

### 从源码构建

```bash
# 准备数据库
createdb linuxdo_relay

# 配置环境变量（参考 .env.example）
export APP_PG_DSN="postgres://..."
export APP_REDIS_ADDR="localhost:6379"
export APP_JWT_SECRET="..."

# 启动后端（自动迁移数据库）
go run ./cmd/server

# 启动前端（开发模式）
cd web && npm install && npm run dev
```

## 环境变量

| 变量名 | 必填 | 说明 |
|--------|------|------|
| `APP_PG_DSN` | **是** | PostgreSQL 连接字符串 |
| `APP_REDIS_ADDR` | **是** | Redis 地址 |
| `APP_JWT_SECRET` | **是** | JWT 密钥（至少 32 字符） |
| `APP_LINUXDO_CLIENT_ID` | 是 | LinuxDo OAuth 客户端 ID |
| `APP_LINUXDO_CLIENT_SECRET` | 是 | LinuxDo OAuth 客户端密钥 |
| `APP_LINUXDO_REDIRECT_URL` | 是 | OAuth 回调地址 |
| `APP_HTTP_LISTEN` | 否 | HTTP 监听地址（默认 `:8080`） |
| `APP_SIGNUP_CREDITS` | 否 | 新用户初始积分（默认 `100`） |

## 使用说明

### 首次登录

1. 访问 http://localhost:8080
2. 点击"使用 LinuxDo 登录"
3. **首个注册用户自动成为管理员**

### API 调用

```bash
curl -X POST https://your-domain.com/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "Hello"}]}'
```

## 目录结构

```
linuxdo-relay/
├── cmd/server/          # 主程序入口
├── internal/
│   ├── auth/           # 认证模块
│   ├── config/         # 配置加载
│   ├── models/         # 数据模型（GORM AutoMigrate）
│   ├── relay/          # API 转发代理
│   ├── server/         # HTTP 路由
│   └── storage/        # 数据库与 Redis
├── web/                # React 前端
├── docker-compose.yml
└── Dockerfile
```

## 许可证

详见 [LICENSE](./LICENSE) 文件。

## 致谢

- [Gin](https://gin-gonic.com/) / [GORM](https://gorm.io/) / [Semi Design](https://semi.design/) / [LinuxDo](https://linux.do/)
