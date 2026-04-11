# emotionalBeach

[📄 中文项目结构文档](./docs/PROJECT_STRUCTURE_CN.md) | [📄 English Project Structure](./docs/PROJECT_STRUCTURE_EN.md)

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/eric-jxl/emotionalbeach?color=blue&label=go&logo=go)
[![build-go-binary](https://github.com/eric-jxl/emotionalbeach/actions/workflows/go-binary-release.yml/badge.svg)](https://github.com/eric-jxl/emotionalbeach/actions/workflows/go-binary-release.yml)
[![Docker Image CI](https://github.com/eric-jxl/emotionalbeach/actions/workflows/docker-image.yml/badge.svg)](https://github.com/eric-jxl/emotionalbeach/actions/workflows/docker-image.yml)

## 📋 项目概览

**emotionalBeach** 是一个基于 Go 语言的后端 API 服务，使用 **Gin** 框架构建，集成了用户认证、好友关系管理、邮件通知等功能。  
项目已完成**全面架构重构**，引入 **Google Wire** 依赖注入、领域驱动分层设计和优雅启停机制，彻底消除了全局可变状态。

**开发环境：** Go v1.24 + Gin v1.10.1 + GORM v1.30.2 + Viper v1.20.1 + Wire v0.7.0

### 🎯 核心功能

| 功能 | 说明 |
|------|------|
| ✅ 用户认证系统 | 用户名/密码登录 + GitHub OAuth 一键登录 |
| ✅ 好友关系管理 | 添加好友（ID / 昵称）、获取好友列表 |
| ✅ JWT Token 认证 | 7 天有效期，Bearer 格式，配置注入密钥 |
| ✅ Swagger 自动注入 Token | 登录页登录后自动写入 `localStorage`，Swagger UI 打开即完成 ApiKeyAuth 授权 |
| ✅ 邮件通知服务 | 多收件人、HTML 内容，配置驱动（无环境变量依赖） |
| ✅ Redis 缓存 | 启动时预热，可通过配置开关 |
| ✅ IP 限流保护 | 令牌桶算法，防止滥用 |
| ✅ Prometheus 指标 | `/metrics` 端点，开箱即用 |
| ✅ Swagger API 文档 | 自动生成，`/swagger/index.html` 访问 |
| ✅ 多数据库支持 | PostgreSQL / MySQL 配置热切换 |
| ✅ 优雅启停 | `SIGINT/SIGTERM` 信号驱动，顺序关闭 HTTP → DB → Redis → Logger |

---

## 🏗️ 架构设计

### 重构亮点

| 项目 | 重构前 | 重构后 |
|------|--------|--------|
| 依赖管理 | 全局变量（`initialize.MainDB`、`global.RedisClient`） | **Google Wire** 编译期依赖注入 |
| 服务生命周期 | `main.go` 手动拼装，关停逻辑分散 | `App.Run()` / `App.shutdown()` 统一管理 |
| 路由注册 | 单文件硬编码所有路由 | 各 Handler 自带 `RegisterRoutes()` 模块注册 |
| 数据访问层 | 包级函数直接依赖全局 DB | `Repository` 接口 + GORM 实现，通过构造注入 |
| 业务逻辑层 | Controller 中混杂业务逻辑 | `Service` 接口 + 纯业务实现，Handler 只做参数绑定和响应 |
| 日志中间件 | 依赖 `initialize.AccessLogger` 全局变量 | `ZapLogger(accessLogger)` 工厂注入 |
| 配置与密钥 | `os.Getenv()` 散落各处 | 统一从 `config.Config` 注入，无隐式环境变量 |

### 分层依赖图

```
main.go
  └── app.InitializeApp(cfg)          ← Wire 编译期生成
        ├── infra/db     → *gorm.DB, *redis.Client
        ├── infra/logger → *Loggers (Sys + Access)
        ├── repository/  → 实现 domain.Repository 接口
        ├── service/     → 实现 domain.Service  接口
        ├── handler/     → 组合 Service，注册路由
        ├── server/      → *gin.Engine, *http.Server
        └── App          → Run() / shutdown()
```

### 目录结构

```
emotionalbeach/
├── main.go                        # 极简入口：加载配置 → Wire → App.Run()
├── config/
│   ├── config.go                  # 配置结构体定义
│   └── config.yaml                # 配置文件（数据库/Redis/日志/服务器）
├── internal/
│   ├── app/
│   │   ├── app.go                 # App 结构体：Run() 优雅启停，shutdown() 顺序关闭
│   │   ├── wire.go                # Wire 注入器声明（wireinject build tag）
│   │   └── wire_gen.go            # Wire 自动生成，勿手动修改
│   ├── domain/                    # 领域契约层（接口定义，零依赖）
│   │   ├── user/interfaces.go     # UserRepository / UserService / UpdateRequest
│   │   └── relation/interfaces.go # RelationRepository / RelationService
│   ├── infra/                     # 基础设施层（Wire Provider）
│   │   ├── db/provider.go         # ProvideDB(*gorm.DB) + ProvideRedis(*redis.Client)
│   │   ├── logger/provider.go     # ProvideLoggers(*Loggers{Sys,Access})
│   │   ├── metrics/               # Prometheus 指标定义 + DB 连接池 Collector
│   │   └── cache/preload.go       # Redis 启动预热
│   ├── repository/                # 数据访问实现层
│   │   ├── user/gorm_repo.go      # 实现 domain/user.Repository（GORM）
│   │   └── relation/gorm_repo.go  # 实现 domain/relation.Repository（GORM）
│   ├── service/                   # 业务逻辑实现层
│   │   ├── user/svc.go            # 实现 domain/user.Service
│   │   ├── relation/svc.go        # 实现 domain/relation.Service
│   │   ├── github/svc.go          # GitHub OAuth 流程封装
│   │   ├── health/svc.go          # 深度健康检查（DB + Redis 探活）
│   │   └── notification/email.go  # SMTP 邮件发送（配置注入，无 os.Getenv）
│   ├── handler/                   # HTTP 表现层（仅参数绑定 + 响应）
│   │   ├── user/                  # handler.go + routes.go
│   │   ├── relation/              # handler.go + routes.go
│   │   ├── github/                # handler.go + routes.go
│   │   ├── health/                # handler.go + routes.go
│   │   └── webhook/               # handler.go + routes.go
│   ├── server/
│   │   ├── router.go              # 组合所有 Handler，装配 *gin.Engine（依赖注入）
│   │   └── server.go              # 构建 *http.Server（超时配置注入）
│   ├── middleware/                # 中间件（JWT、限流、结构化日志、Prometheus）
│   ├── models/                    # GORM 数据模型（UserBasic、Relation）
│   ├── common/                    # 纯函数工具（MD5、手机号校验）
│   ├── global/                    # HTTP 统一响应工具（RespJson / Success / Error）
│   └── templates/                 # 嵌入式前端模板（登录页、Swagger UI、静态资源）
├── docs/                          # Swagger 自动生成文档
└── monitoring/                    # Prometheus + Grafana 监控配置
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
# 下载依赖
go mod download

# 安装 Wire 代码生成工具
go install github.com/google/wire/cmd/wire@latest

# 安装 Swagger 文档生成工具
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. 配置文件

编辑 `config/config.yaml`，关键配置项：

```yaml
server:
  port: 8080
  jwtSecret: "your-secret-key"
  clientID: "github-client-id"       # GitHub OAuth
  clientSecret: "github-client-secret"
  enableRedis: true
  shutdownTimeoutSec: 10

databases:
  main:
    type: postgres                    # 或 mysql
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
# 首次运行，自动建表后退出
go run main.go -migrate
```

### 4. 运行服务

```bash
# 直接运行
go run main.go

# 编译后运行
make all
./emotionalBeach

# Docker Compose
docker compose up -d
```

### 5. 重新生成 Wire 注入代码

当新增 Provider 或修改依赖图时，重新生成 `wire_gen.go`：

```bash
wire gen ./internal/app/...
```

### 6. 生成 Swagger 文档

```bash
make gen
# 或
swag init -o ./docs -g main.go
```

### 7. 常用 Make 命令

```bash
make all   # 编译 + upx 压缩
make gen   # 生成 Swagger 文档
```

---

## 🔑 Swagger 自动 Token 注入

登录页（`/`）支持**一次登录，Swagger 自动授权**，无需手动粘贴 Token：

```
打开登录页 /
  ↓ 输入账号密码，点击登录
POST /login → 返回 JWT Token
  ↓
Token 自动写入 localStorage（键名：eb_token）
  ↓
跳转 /swagger/index.html
  ↓
Swagger UI 初始化完成（onComplete）
  ↓
ui.preauthorizeApiKey('ApiKeyAuth', 'Bearer <token>')
  ↓
顶部绿色提示条：🔑 已自动注入登录 Token ✅
所有受保护接口可直接调用，无需手动填写 Authorization
```

> Token 通过 `persistAuthorization: true` 持久化，刷新页面不丢失。  
> 若需切换用户，点击 Swagger UI 右上角 **Authorize** 按钮重新授权即可。

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
| GET | `/swagger/*` | Swagger 交互式文档（含自动 Token 注入） |

---

## 🔐 安全特性

| 特性 | 实现方式 |
|------|----------|
| 密码加密 | MD5 + 随机盐值（`common.SaltPassWord`） |
| JWT 认证 | HS256，7 天有效期，密钥从配置注入（无全局写入竞态） |
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
| Lumberjack | v2.2.1 | 日志轮转 |
| Prometheus | latest | 可观测性指标 |
| Swagger (swag) | v1.16.6 | API 文档生成 |

---

## 📖 详细文档

- [📄 中文项目结构文档](./docs/PROJECT_STRUCTURE_CN.md)
- [📄 英文项目结构文档](./docs/PROJECT_STRUCTURE_EN.md)
- [📋 中文 API 接口文档](./docs/API_DOCS_CN.md)
- [📋 英文 API 接口文档](./docs/API_DOCS_EN.md)
- [📚 文档索引导航](./docs/INDEX.md)
- [📋 快速参考](./docs/QUICK_REFERENCE.md)

---

## 📞 贡献指南

欢迎提交 Issue 和 Pull Request！

---

## 📄 许可证

Apache License 2.0
