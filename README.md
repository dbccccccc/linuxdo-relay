# linuxdo-relay

一个独立的 API 转发服务，基于 Go + Gin + PostgreSQL + Redis

前端控制台使用 React + Vite + Semi UI，提供个人中心（配额用量、调用/操作日志、API Key 重置、每日签到）和管理员功能（用户、渠道、配额规则、积分规则、日志、统计面板）。

## 功能特性

- ✅ **LinuxDo OAuth 登录**：集成 LinuxDo 账号体系
- ✅ **双重认证**：支持 JWT Token（Web）和 API Key（程序调用）
- ✅ **配额限流**：基于用户等级和模型前缀的灵活限流策略
- ✅ **积分系统**：按模型计费，支持预扣和失败退款
- ✅ **每日签到**：积分奖励，连续签到统计，余额衰减机制
- ✅ **渠道管理**：多上游渠道，模型唯一性约束
- ✅ **完整日志**：API 调用、登录、操作、积分交易记录
- ✅ **管理后台**：用户管理、渠道配置、规则设置、数据统计

## 环境要求

- Go 1.21+
- PostgreSQL 13+
- Redis 6+
- Node.js 16+（前端构建）

## 快速开始

### 方式一：使用预编译二进制（最简单）

1. **下载最新版本**

访问 [Releases 页面](https://github.com/dbccccccc/linuxdo-relay/releases) 下载适合你系统的二进制文件：
- `linuxdo-relay-linux-amd64` - Linux (x86_64)
- `linuxdo-relay-linux-arm64` - Linux (ARM64)
- `linuxdo-relay-darwin-amd64` - macOS (Intel)
- `linuxdo-relay-darwin-arm64` - macOS (Apple Silicon)
- `linuxdo-relay-windows-amd64.exe` - Windows (x86_64)

2. **下载数据库迁移文件**
```bash
# 下载并解压 migrations.tar.gz
wget https://github.com/dbccccccc/linuxdo-relay/releases/latest/download/migrations.tar.gz
tar -xzf migrations.tar.gz
```

3. **配置环境变量**（参考下方配置说明）

4. **初始化数据库**
```bash
psql -d linuxdo_relay -f migrations/001_init.sql
psql -d linuxdo_relay -f migrations/002_logs.sql
psql -d linuxdo_relay -f migrations/003_credits.sql
psql -d linuxdo_relay -f migrations/004_check_in.sql
```

5. **运行程序**
```bash
# Linux/macOS
chmod +x linuxdo-relay-linux-amd64
./linuxdo-relay-linux-amd64

# Windows
.\linuxdo-relay-windows-amd64.exe
```

### 方式二：Docker Compose（推荐生产环境）

1. **克隆项目或下载 docker-compose.yml**
```bash
git clone https://github.com/dbccccccc/linuxdo-relay.git
cd linuxdo-relay

# 或直接下载 docker-compose.yml
wget https://github.com/dbccccccc/linuxdo-relay/releases/latest/download/docker-compose.yml
```

2. **配置环境变量**
```bash
# 编辑 docker-compose.yml，修改以下关键配置：
# - APP_LINUXDO_CLIENT_ID: 你的 LinuxDo OAuth 客户端 ID
# - APP_LINUXDO_CLIENT_SECRET: 你的 LinuxDo OAuth 客户端密钥
# - APP_LINUXDO_REDIRECT_URL: OAuth 回调地址
# - APP_JWT_SECRET: 修改为强密码
```

3. **启动服务**
```bash
docker-compose up -d
```

4. **初始化数据库**
```bash
# 进入 postgres 容器执行迁移
docker-compose exec postgres psql -U relay -d linuxdo_relay

# 在 psql 中执行迁移脚本
\i /docker-entrypoint-initdb.d/001_init.sql
\i /docker-entrypoint-initdb.d/002_logs.sql
\i /docker-entrypoint-initdb.d/003_credits.sql
\i /docker-entrypoint-initdb.d/004_check_in.sql
\q
```

或者直接在宿主机执行：
```bash
cat migrations/*.sql | docker-compose exec -T postgres psql -U relay -d linuxdo_relay
```

5. **访问服务**
- 后端 API：http://localhost:8080
- 健康检查：http://localhost:8080/healthz

### 方式三：使用 Docker 镜像

```bash
# 从 Docker Hub 拉取
docker pull dbccccccc/linuxdo-relay:latest

# 或从 GitHub Container Registry 拉取
docker pull ghcr.io/dbccccccc/linuxdo-relay:latest

# 运行容器
docker run -d \
  -p 8080:8080 \
  -e APP_PG_DSN="postgres://user:pass@host:5432/linuxdo_relay" \
  -e APP_REDIS_ADDR="redis:6379" \
  -e APP_JWT_SECRET="your-secret" \
  -e APP_LINUXDO_CLIENT_ID="your-client-id" \
  -e APP_LINUXDO_CLIENT_SECRET="your-client-secret" \
  -e APP_LINUXDO_REDIRECT_URL="http://localhost:8080/auth/linuxdo/callback" \
  dbccccccc/linuxdo-relay:latest
```

### 方式四：从源码构建（开发者）

#### 1. 准备数据库

**启动 PostgreSQL**
```bash
# 创建数据库
createdb linuxdo_relay

# 执行迁移
psql -d linuxdo_relay -f migrations/001_init.sql
psql -d linuxdo_relay -f migrations/002_logs.sql
psql -d linuxdo_relay -f migrations/003_credits.sql
psql -d linuxdo_relay -f migrations/004_check_in.sql
```

**启动 Redis**
```bash
redis-server
```

#### 2. 配置环境变量

创建 `.env` 文件或导出环境变量：

```bash
export APP_HTTP_LISTEN=":8080"
export APP_PG_DSN="postgres://user:password@localhost:5432/linuxdo_relay?sslmode=disable"
export APP_REDIS_ADDR="localhost:6379"
export APP_REDIS_PASSWORD=""
export APP_JWT_SECRET="your-very-secure-secret-key-change-me"
export APP_SIGNUP_CREDITS="100"
export APP_DEFAULT_MODEL_CREDIT_COST="1"

# LinuxDo OAuth 配置（必填）
export APP_LINUXDO_CLIENT_ID="your-client-id"
export APP_LINUXDO_CLIENT_SECRET="your-client-secret"
export APP_LINUXDO_REDIRECT_URL="http://localhost:8080/auth/linuxdo/callback"

# 可选：自定义 LinuxDo OAuth 端点
# export APP_LINUXDO_AUTH_URL="https://linux.do/oauth2/authorize"
# export APP_LINUXDO_TOKEN_URL="https://linux.do/oauth2/token"
# export APP_LINUXDO_USERINFO_URL="https://linux.do/api/user"
```

#### 3. 启动后端

```bash
go mod download
go run ./cmd/server
```

#### 4. 启动前端（开发模式）

```bash
cd web
npm install
npm run dev
```

前端开发服务器：http://localhost:5173

## 配置说明

### 环境变量

| 变量名 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `APP_HTTP_LISTEN` | 否 | `:8080` | HTTP 监听地址 |
| `APP_PG_DSN` | 是 | - | PostgreSQL 连接字符串 |
| `APP_REDIS_ADDR` | 否 | `localhost:6379` | Redis 地址 |
| `APP_REDIS_PASSWORD` | 否 | `""` | Redis 密码 |
| `APP_JWT_SECRET` | 是 | - | JWT 签名密钥（至少 32 位） |
| `APP_SIGNUP_CREDITS` | 否 | `100` | 新用户初始积分 |
| `APP_DEFAULT_MODEL_CREDIT_COST` | 否 | `1` | 默认模型扣费 |
| `APP_LINUXDO_CLIENT_ID` | 是 | - | LinuxDo OAuth 客户端 ID |
| `APP_LINUXDO_CLIENT_SECRET` | 是 | - | LinuxDo OAuth 客户端密钥 |
| `APP_LINUXDO_REDIRECT_URL` | 是 | - | OAuth 回调地址 |
| `APP_LINUXDO_AUTH_URL` | 否 | 见配置文件 | LinuxDo 授权端点 |
| `APP_LINUXDO_TOKEN_URL` | 否 | 见配置文件 | LinuxDo Token 端点 |
| `APP_LINUXDO_USERINFO_URL` | 否 | 见配置文件 | LinuxDo 用户信息端点 |

### LinuxDo OAuth 配置

1. 访问 LinuxDo 开发者设置：https://linux.do/admin/api/keys
2. 创建新的 OAuth2 应用
3. 设置回调地址：`http://your-domain:8080/auth/linuxdo/callback`
4. 获取 Client ID 和 Client Secret
5. 将凭证配置到环境变量中

## 部署指南

### 生产环境部署（推荐配置）

#### 1. 使用 Docker Compose

修改 `docker-compose.yml`：

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:14
    restart: unless-stopped
    environment:
      POSTGRES_USER: relay
      POSTGRES_PASSWORD: <strong-password>
      POSTGRES_DB: linuxdo_relay
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    networks:
      - linuxdo-network

  redis:
    image: redis:7
    restart: unless-stopped
    command: redis-server --requirepass <redis-password>
    volumes:
      - redis-data:/data
    networks:
      - linuxdo-network

  linuxdo-relay:
    build: .
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      APP_HTTP_LISTEN: ":8080"
      APP_PG_DSN: "postgres://relay:<strong-password>@postgres:5432/linuxdo_relay?sslmode=disable"
      APP_REDIS_ADDR: "redis:6379"
      APP_REDIS_PASSWORD: "<redis-password>"
      APP_JWT_SECRET: "<your-jwt-secret>"
      APP_LINUXDO_CLIENT_ID: "${LINUXDO_CLIENT_ID}"
      APP_LINUXDO_CLIENT_SECRET: "${LINUXDO_CLIENT_SECRET}"
      APP_LINUXDO_REDIRECT_URL: "https://your-domain.com/auth/linuxdo/callback"
    depends_on:
      - postgres
      - redis
    networks:
      - linuxdo-network

volumes:
  postgres-data:
  redis-data:

networks:
  linuxdo-network:
```

创建 `Dockerfile`：

```dockerfile
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM node:18-alpine AS frontend-builder
WORKDIR /app
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=backend-builder /server .
COPY --from=frontend-builder /app/dist ./web/dist
COPY migrations ./migrations
EXPOSE 8080
CMD ["./server"]
```

#### 2. 配置 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # 前端静态文件
    location / {
        root /path/to/linuxdo-relay/web/dist;
        try_files $uri $uri/ /index.html;
    }

    # 后端 API
    location /api/ {
        proxy_pass http://localhost:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # OAuth 回调
    location /auth/ {
        proxy_pass http://localhost:8080/auth/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 健康检查
    location /healthz {
        proxy_pass http://localhost:8080/healthz;
    }
}
```

#### 3. 部署步骤

```bash
# 1. 拉取代码
git clone https://github.com/yourusername/linuxdo-relay.git
cd linuxdo-relay

# 2. 配置环境变量
cp .env.example .env
nano .env  # 编辑配置

# 3. 构建并启动
docker-compose up -d --build

# 4. 初始化数据库
cat migrations/*.sql | docker-compose exec -T postgres psql -U relay -d linuxdo_relay

# 5. 检查服务状态
docker-compose ps
curl http://localhost:8080/healthz

# 6. 查看日志
docker-compose logs -f linuxdo-relay
```

### 数据库备份

```bash
# 备份
docker-compose exec postgres pg_dump -U relay linuxdo_relay > backup.sql

# 恢复
cat backup.sql | docker-compose exec -T postgres psql -U relay -d linuxdo_relay
```

### 监控与维护

**查看日志**
```bash
docker-compose logs -f linuxdo-relay
docker-compose logs -f postgres
docker-compose logs -f redis
```

**重启服务**
```bash
docker-compose restart linuxdo-relay
```

**更新部署**
```bash
git pull
docker-compose up -d --build
```

## 使用说明

### 首次登录

1. 访问前端地址（默认 http://localhost:8080 或你的域名）
2. 点击"使用 LinuxDo 登录"
3. 授权后自动创建账户
4. **首个注册用户自动成为管理员**

### 生成 API Key

1. 登录后进入"个人中心"
2. 点击"生成 / 重置 API Key"
3. 复制并保存 API Key（仅显示一次）

### 调用 API

```bash
curl -X POST https://your-domain.com/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

### 管理员功能

详见 [ADMIN_GUIDE.md](./ADMIN_GUIDE.md)

- 用户管理：调整角色、等级、状态
- 渠道管理：配置上游 API
- 配额规则：设置限流策略
- 积分规则：配置模型价格
- 日志查询：查看系统日志
- 数据统计：监控使用情况

## LinuxDo OAuth 集成

根据 LinuxDo 官方文档（LinuxDo Connect），本项目预期实现标准 OAuth2 授权码流程：

1. 用户访问 `/auth/linuxdo/login`，服务端构造 `state` 并重定向到 LinuxDo 授权页；
2. 用户在 LinuxDo 授权后回调到 `/auth/linuxdo/callback`，携带 `code` 与 `state`；
3. 服务端校验 `state`，并通过 `code` 去 LinuxDo 的 Token 端点换取 `access_token`；
4. 使用 `access_token` 调用 LinuxDo 用户信息接口，获取用户唯一 ID 与用户名；
5. 在本地 `users` 表中写入/更新用户记录：
   - 若数据库当前无任何用户，则该用户自动设为 `admin`；
   - 否则新用户为普通 `user`，默认等级为 1，状态为 `normal`；
6. 生成本地 JWT，作为后续 API 调用的认证凭证。

> 相关逻辑已在 `internal/auth` 中实现，你只需要在环境变量中配置正确的客户端 ID/Secret、重定向地址等参数即可正常工作。


## 限额系统

- PostgreSQL 中的 `quota_rules` 表存储配额规则：
  - 维度：`level`（用户等级） + `model_pattern`（模型前缀）；
  - 字段：`max_requests`（窗口内最大请求次数）、`window_seconds`（窗口长度秒）。
- 管理员可以通过 `/admin/quota_rules` 接口管理规则：
  - `GET /admin/quota_rules`：列表
  - `POST /admin/quota_rules`：新增
  - `PUT /admin/quota_rules/:id`：修改
  - `DELETE /admin/quota_rules/:id`：删除
- Redis 使用固定时间窗口计数实现限流：
  - key 形如：`quota:{user_id}:{level}:{model_pattern}:{bucket}`；
  - 每次请求前执行 `INCR`，首次设置 `EXPIRE`；
  - 超出 `max_requests` 则返回 HTTP 429，错误码为 `quota_exceeded`；
- 所有限额均基于“请求次数”，不基于 token 数量；后续 token 使用仅用于展示/统计，不参与限流判断。

## 积分系统

- `users` 表新增 `credits` 字段，积分永远是整数。
- `model_credit_rules` 按模型前缀设定扣费，所有等级统一价格；若未匹配则使用 `APP_DEFAULT_MODEL_CREDIT_COST`。
- 请求进入转发逻辑前先预扣积分，若上游返回非 2xx 或发生代理错误则自动退款。
- 变动记录写入 `credit_transactions`，方便审计及前端展示。
- 管理员可通过 `/admin/model_credit_rules` 维护价格，并通过 `/admin/users/:id/credits` 为用户手动充值/扣减。
- 新注册用户会获得 `APP_SIGNUP_CREDITS` 设定的初始积分，后续可扩展签到等补充渠道。

## Web 控制台

前端位于 `web/` 目录，默认通过浏览器访问后端同源接口。

```bash
cd web
npm install
npm run dev   # http://localhost:5173

# 生产构建
npm run build
```

生产环境可以使用 Nginx/静态服务器托管 `web/dist`，并将流量反向代理到后端 `linuxdo-relay` 服务。

### 控制台功能

- **个人中心 /me**：展示账户信息、生成 API Key、查看配额使用情况、最近 API 调用及操作日志。
- **管理员**：
  - 用户管理：变更角色/等级/状态。
  - 渠道管理：增删改查渠道，限制模型唯一归属。
  - 配额规则：维护 level+model 前缀的限额策略。
  - 日志中心：分页查看 API 调用日志与登录日志。
  - 全局统计：总用户/总调用/24h 活跃用户概览。

所有管理员入口都需持有管理员角色（首次注册用户自动成为管理员）。

## 测试

**后端测试**
```bash
go test ./... -v
```

**前端测试**
```bash
cd web
npm run test
```

## 故障排查

### 常见问题

1. **无法连接数据库**
   - 检查 `APP_PG_DSN` 配置
   - 确认 PostgreSQL 服务运行中
   - 验证数据库用户权限

2. **OAuth 登录失败**
   - 检查 LinuxDo Client ID/Secret
   - 确认回调地址配置正确
   - 查看后端日志错误信息

3. **API 调用返回 401**
   - 确认 API Key 正确
   - 检查用户状态是否为 `normal`
   - 验证 JWT Token 未过期

4. **配额限制不生效**
   - 检查 Redis 连接
   - 确认配额规则已配置
   - 验证用户等级匹配

### 日志调试

```bash
# 查看应用日志
docker-compose logs -f linuxdo-relay

# 查看数据库日志
docker-compose logs -f postgres

# 查看 Redis 日志
docker-compose logs -f redis
```

## 发布流程

项目使用 GitHub Actions 自动化构建和发布。

### 创建新版本

1. **打标签并推送**
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. **自动构建**
GitHub Actions 将自动：
- 构建多平台二进制文件（Linux/Windows/macOS，amd64/arm64）
- 构建前端静态文件
- 构建并推送 Docker 镜像到 Docker Hub 和 GHCR
- 创建 GitHub Release 并上传所有构建产物

3. **配置 Docker Hub（首次）**

在 GitHub 仓库设置中添加 Secrets：
- `DOCKER_USERNAME`: Docker Hub 用户名
- `DOCKER_PASSWORD`: Docker Hub 访问令牌

### 版本号规范

遵循语义化版本 (Semantic Versioning)：
- `v1.0.0` - 主版本.次版本.修订号
- `v1.0.0-beta.1` - 预发布版本
- `v1.0.0-rc.1` - 候选版本

## 开发指南

### 目录结构

```
linuxdo-relay/
├── cmd/server/          # 主程序入口
├── internal/
│   ├── auth/           # 认证模块（JWT、API Key、OAuth）
│   ├── config/         # 配置加载
│   ├── models/         # 数据模型
│   ├── relay/          # API 转发代理
│   ├── server/         # HTTP 服务与路由
│   └── storage/        # 数据库与 Redis
├── migrations/         # 数据库迁移脚本
├── web/               # React 前端
│   └── src/
│       └── modules/
│           ├── auth/   # 登录与认证
│           ├── me/     # 个人中心
│           └── admin/  # 管理后台
├── docker-compose.yml  # Docker 编排
├── README.md          # 项目文档
└── ADMIN_GUIDE.md     # 管理员指南
```

### 添加新功能

1. **添加数据模型**：在 `internal/models/` 创建新模型
2. **创建迁移**：在 `migrations/` 添加 SQL 脚本
3. **实现路由**：在 `internal/server/` 添加路由处理
4. **前端页面**：在 `web/src/modules/` 添加组件
5. **编写测试**：添加单元测试和集成测试

### 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 安全建议

- ✅ 使用强密码配置 `APP_JWT_SECRET`（至少 32 字符）
- ✅ 生产环境启用 PostgreSQL SSL 连接
- ✅ Redis 设置密码保护
- ✅ 使用 HTTPS 部署（配置 SSL 证书）
- ✅ 定期备份数据库
- ✅ 限制管理员账户数量
- ✅ 定期审查用户权限和日志
- ✅ 及时更新依赖版本

## 性能优化

- 配置 PostgreSQL 连接池
- 启用 Redis 持久化（AOF/RDB）
- 使用 CDN 加速前端资源
- 配置 Nginx 缓存静态文件
- 监控数据库慢查询
- 定期清理历史日志数据

## 许可证

详见 [LICENSE](./LICENSE) 文件。

## 支持

- 📖 [项目文档](./README.md)
- 📋 [管理员指南](./ADMIN_GUIDE.md)
- 🐛 [问题反馈](https://github.com/yourusername/linuxdo-relay/issues)
- 💬 [讨论区](https://github.com/yourusername/linuxdo-relay/discussions)

## CI/CD

项目配置了以下 GitHub Actions 工作流：

### Build and Test
- **触发条件**: Push 到 main/develop 分支，或 PR
- **功能**:
  - 后端单元测试和代码覆盖率
  - 前端测试和构建验证
  - Go 代码 lint 检查
  - Docker 镜像构建测试

### Release
- **触发条件**: 推送版本标签 (v*)
- **功能**:
  - 构建多平台二进制文件
  - 构建前端静态文件
  - 构建并推送 Docker 镜像
  - 创建 GitHub Release
  - 生成校验和文件

### 徽章

[![Build](https://github.com/dbccccccc/linuxdo-relay/actions/workflows/build.yml/badge.svg)](https://github.com/dbccccccc/linuxdo-relay/actions/workflows/build.yml)
[![Release](https://github.com/dbccccccc/linuxdo-relay/actions/workflows/release.yml/badge.svg)](https://github.com/dbccccccc/linuxdo-relay/actions/workflows/release.yml)
[![Docker](https://img.shields.io/docker/v/dbccccccc/linuxdo-relay?label=docker&sort=semver)](https://hub.docker.com/r/dbccccccc/linuxdo-relay)
[![License](https://img.shields.io/github/license/dbccccccc/linuxdo-relay)](./LICENSE)

## 致谢

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Semi Design](https://semi.design/)
- [LinuxDo Community](https://linux.do/)

