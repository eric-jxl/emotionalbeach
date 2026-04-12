# emotionalBeach

[📄 中文项目结构文档](./docs/PROJECT_STRUCTURE_CN.md) | [📄 English Project Structure](./docs/PROJECT_STRUCTURE_EN.md)

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/eric-jxl/emotionalbeach?color=blue&label=go&logo=go)
[![build-go-binary](https://github.com/eric-jxl/emotionalbeach/actions/workflows/go-binary-release.yml/badge.svg)](https://github.com/eric-jxl/emotionalbeach/actions/workflows/go-binary-release.yml)
[![Docker Image CI](https://github.com/eric-jxl/emotionalbeach/actions/workflows/docker-image.yml/badge.svg)](https://github.com/eric-jxl/emotionalbeach/actions/workflows/docker-image.yml)

## 📋 项目概览

**emotionalBeach** 是一个基于 Go 语言的后端 API 服务，使用 **Gin** 框架构建，集成了用户认证、好友关系管理、邮件通知等功能。

**开发环境：** Go v1.24 + Gin v1.10.1 + GORM v1.30.2 + Viper v1.20.1 + Wire v0.7.0

### 🎯 核心功能

| 功能 | 说明 |
|------|------|
| ✅ 用户认证系统 | 用户名/密码登录 + GitHub OAuth 一键登录 |
| ✅ 好友关系管理 | 添加好友（ID / 昵称）、获取好友列表 |
| ✅ JWT Token 认证 | 7 天有效期，Bearer 格式，配置注入密钥 |
| ✅ 内嵌登录页 | 密码 / 验证码 / 扫码三种登录方式 UI，登录后自动跳转 Swagger |
| ✅ Swagger 自动注入 Token | 登录后自动写入 `localStorage`，Swagger UI 打开即完成 ApiKeyAuth 授权 |
| ✅ 邮件通知服务 | 多收件人、HTML 内容，配置驱动（无环境变量依赖） |
| ✅ Redis 缓存 | 启动时预热，可通过配置开关 |
| ✅ IP 限流保护 | 令牌桶算法，防止滥用 |
| ✅ Prometheus 指标 | `/metrics` 端点，开箱即用 |
| ✅ Swagger API 文档 | 自动生成，`/swagger/index.html` 访问 |
| ✅ 多数据库支持 | PostgreSQL / MySQL 配置热切换 |
| ✅ 优雅启停 | `SIGINT/SIGTERM` 信号驱动，Wire cleanup 顺序关闭 HTTP → DB → Redis → Logger |

---

## 🏗️ 架构设计

### 分层依赖图

```
main.go
  └── di.InitializeApp(cfg)           ← Wire 编译期注入（5 个 Provider）
        ├── infra.Provider            → *gorm.DB, *redis.Client, *Loggers
        ├── dao.Provider              → dao.Dao（统一数据访问接口）
        ├── service.Provider          → *service.Service（所有业务能力）
        ├── server.New                → *http.Server（含所有路由和 Handler）
        └── di.NewApp                 → *App → Run()
```

### Wire 注入（极简 5 Provider）

```go
// internal/di/wire.go
func InitializeApp(cfg *config.Config) (*App, func(), error) {
    panic(wire.Build(infra.Provider, dao.Provider, service.Provider, server.New, NewApp))
}
```

### 统一 API 响应格式

所有接口统一使用 `internal/common` 包中的**三个**函数封装返回：

| 函数 | 场景 | HTTP 状态 | `code` 字段 |
|------|------|-----------|-------------|
| `common.Success(c, data)` | 请求成功 | 200 | `0` |
| `common.Fail(c, httpStatus, msg)` | 业务失败（参数错误、鉴权失败等 4xx） | 传入值（4xx） | 同 HTTP 状态码 |
| `common.ServerError(c, msg)` | 服务内部错误（5xx） | 500 | `500` |

```json
// ✅ 成功 — common.Success(c, data)
{ "code": 200, "message": "success", "data": { ... } }

// ⚠️ 业务失败 — common.Fail(c, 400, "msg")
{ "code": 400, "message": "手机号必须为 11 位" }

// ❌ 服务错误 — common.ServerError(c, "msg")
{ "code": 500, "message": "修改信息失败: ..." }
```

### 目录结构

```
emotionalbeach/
├── main.go                          # 极简入口：加载配置 → Wire → App.Run()
├── config/config.go + config.yaml   # 配置结构体 + 配置文件
├── internal/
│   ├── di/
│   │   ├── wire.go                  # Wire 注入器（wireinject build tag）
│   │   ├── wire_gen.go              # Wire 自动生成，勿手动修改
│   │   └── app.go                   # App 结构体：Run() 优雅启停
│   ├── infra/                       # 基础设施（单一扁平包）
│   │   ├── infra.go                 # Provider = wire.NewSet(dbSet, loggerSet)
│   │   ├── db.go                    # ProvideDB + ProvideRedis
│   │   ├── logger.go                # ProvideLoggers + Loggers 类型
│   │   ├── metrics.go               # Prometheus 指标定义 + DB Pool Collector
│   │   └── startup.go               # AutoMigrate / CachePreload / RegisterCollectors
│   ├── dao/                         # 统一数据访问层
│   │   ├── dao.go                   # Dao 接口 + dao struct + Provider
│   │   ├── user.go                  # User GORM 查询实现
│   │   └── relation.go              # Relation GORM 查询实现
│   ├── service/                     # 统一业务逻辑层
│   │   ├── service.go               # Service struct + New() + Provider
│   │   ├── user.go                  # 用户业务方法
│   │   ├── relation.go              # 关系业务方法
│   │   ├── github.go                # GitHub OAuth 方法
│   │   ├── health.go                # 健康检查方法 + 类型定义
│   │   └── email.go                 # 邮件发送方法
│   ├── server/                      # HTTP 层（arip-samp 风格，package-level svc）
│   │   ├── server.go                # var svc + New() + initRouter()
│   │   ├── user.go                  # 用户 Handler 函数
│   │   ├── relation.go              # 关系 Handler 函数
│   │   ├── github.go                # GitHub Handler 函数
│   │   ├── health.go                # 健康检查 Handler 函数
│   │   └── webhook.go               # Webhook/Email Handler 函数
│   ├── middleware/                  # 中间件（JWT、限流、日志、Prometheus）
│   ├── models/                      # GORM 数据模型（UserBasic、Relation）
│   ├── common/                      # 纯函数工具包
│   │   ├── md5.go                   # MD5、密码加盐、手机号校验
│   │   └── response.go              # 统一 HTTP 响应：Success / Fail / ServerError
│   └── templates/                   # 嵌入式前端（登录页、Swagger UI、静态资源）
├── docs/                            # Swagger 自动生成文档
└── monitoring/                      # Prometheus + Grafana 监控配置
```

---

## 🚀 快速启动

```bash
# 推荐：Docker Compose 一键启动
docker compose up -d

# 拉取最新镜像
docker pull ghcr.io/eric-jxl/emotionalbeach:latest
```

> [!TIP]
> - 启动前需配置 `config/config.yaml`（数据库、Redis、端口等）
> - 所有敏感配置支持**环境变量覆盖**（`SERVER_JWTSECRET`、`DATABASES_*` 等）
> - 默认监听端口 **8080**

---

## 🛠️ 开发指南

### 1. 环境准备

```bash
go mod download
go install github.com/google/wire/cmd/wire@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. 配置文件

编辑 `config/config.yaml`，关键配置项：

```yaml
server:
  port: 8080
  jwtSecret: "your-secret-key"
  clientID: "github-client-id"
  clientSecret: "github-client-secret"
  enableRedis: true
  shutdownTimeoutSec: 10

databases:
  main:
    type: postgres          # 或 mysql
    host: localhost
    port: 5432
    user: postgres
    password: password
    dbname: emotionalbeach
default_database: main

mail:
  smtpUser: "xxx@qq.com"
  smtpPassword: "your-smtp-password"
  mailFrom: "xxx@qq.com"
```

### 3. 数据库迁移

```bash
go run main.go -migrate
```

### 4. 运行服务

```bash
go run main.go      # 直接运行
make all            # 编译 + upx 压缩
docker compose up -d
```

### 5. 重新生成 Wire 注入代码

```bash
wire gen ./internal/di/...
```

### 6. 生成 Swagger 文档

```bash
make gen
# 或
swag init -o ./docs -g main.go
```

---

## 🔑 登录 & Swagger 自动 Token 注入

服务内置了一个完整的登录页面（`/`），支持三种登录方式：

| 方式 | 说明 |
|------|------|
| 密码登录 | 输入手机号 / 邮箱 + 密码，点击登录 |
| 验证码登录 | 手机号 + 短信验证码（待接入短信服务） |
| 扫码登录 | 预留入口（待实现） |

登录成功后的完整流程：

```
打开 /  →  输入账号密码  →  POST /login 返回 JWT
  → Token 自动写入 localStorage  →  跳转 /swagger/index.html
  → Swagger UI 自动完成 ApiKeyAuth 授权（顶部绿色提示）
  → 所有受保护接口可直接调用，无需手动填写 Authorization
```

---

## 📡 API 接口概览

### 👤 用户（User）

| 方法 | 端点 | 认证 | 功能 |
|------|------|:----:|------|
| POST | `/register` | ❌ | 用户注册（表单） |
| POST | `/login` | ❌ | 密码登录，返回 JWT |
| GET | `/login/github` | ❌ | GitHub OAuth 跳转 |
| GET | `/callback` | ❌ | GitHub OAuth 回调 |
| GET | `/v1/user/list` | ✅ | 获取所有用户 |
| GET | `/v1/user/condition` | ✅ | 条件查询（id / phone / email） |
| POST | `/v1/user/update` | ✅ | 更新用户信息 |
| DELETE | `/v1/user/delete` | ✅ | 软删除用户 |

### 👥 好友关系（Relation）

| 方法 | 端点 | 认证 | 功能 |
|------|------|:----:|------|
| POST | `/v1/relation/list` | ✅ | 获取好友列表 |
| POST | `/v1/relation/add` | ✅ | 添加好友（ID 或昵称） |

### 📧 Webhook / 通知

| 方法 | 端点 | 认证 | 功能 |
|------|------|:----:|------|
| POST | `/v1/api/webhook` | ✅ | 异步发送邮件通知 |

### 🏥 系统

| 方法 | 端点 | 功能 |
|------|------|------|
| GET | `/ping` | 快速存活检查 |
| GET | `/health` | 深度健康检查（DB + Redis 状态） |
| GET | `/metrics` | Prometheus 指标抓取 |
| GET | `/swagger/*` | Swagger 交互式文档 |

---

## 🔐 安全特性

| 特性 | 实现方式 |
|------|----------|
| 密码加密 | MD5 + 随机盐值（`common.SaltPassWord`） |
| JWT 认证 | HS256，7 天有效期，密钥从配置注入 |
| IP 限流 | 令牌桶，10 秒内最多 5 次（可配置） |
| 结构化日志 | Zap + Lumberjack 滚动，控制台彩色 + JSON 文件双输出 |
| 请求追踪 | 每个请求自动注入 `X-Request-Id` |

---

## 📊 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Gin | v1.10.1 | HTTP 框架 |
| GORM | v1.30.2 | ORM |
| Google Wire | v0.7.0 | 编译期依赖注入 |
| JWT | v5.3.0 | Token 认证 |
| Redis (go-redis) | v9.17.2 | 缓存层 |
| Viper | v1.20.1 | 配置管理 |
| Zap | v1.27.1 | 结构化日志 |
| Prometheus | latest | 可观测性指标 |
| Swagger (swag) | v1.16.6 | API 文档生成 |

---

## 📖 详细文档

- [📄 中文项目结构文档](./docs/PROJECT_STRUCTURE_CN.md)
- [📄 英文项目结构文档](./docs/PROJECT_STRUCTURE_EN.md)
- [📋 中文 API 接口文档](./docs/API_DOCS_CN.md)
- [📋 英文 API 接口文档](./docs/API_DOCS_EN.md)

---

## 📄 许可证

Apache License 2.0
