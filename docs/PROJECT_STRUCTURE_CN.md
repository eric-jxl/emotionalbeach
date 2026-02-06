# 情感沙滩 (emotionalBeach) - 项目结构文档

## 📋 项目概述

**emotionalBeach** 是一个基于 Go 语言的后端 API 服务，使用 Gin 框架构建，集成了用户认证、好友关系管理、邮件通知等功能。该项目支持 GitHub OAuth 认证、JWT Token 验证、Redis 缓存、多数据库支持（PostgreSQL/MySQL）。

**开发环境：** Go v1.23.7 + Gin v1.10.1 + GORM v1.30.2 + Viper v1.20.1

---

## 📁 目录结构详解

### 根目录文件

| 文件/目录 | 作用描述 |
|---------|--------|
| **main.go** | 项目入口文件，初始化日志、数据库、Redis，启动 HTTP 服务器 |
| **go.mod** | Go 模块依赖声明文件 |
| **go.sum** | Go 模块依赖版本锁文件 |
| **Makefile** | 编译快捷命令（如 `make all`, `make gen`） |
| **Dockerfile** | Docker 镜像构建文件 |
| **docker-compose.yml** | Docker Compose 配置（用于本地开发环境）|
| **deploy.sh** | 部署脚本 |
| **entrypoint.sh** | Docker 容器启动脚本 |
| **README.md** | 项目说明文档 |
| **LICENSE** | 开源许可证（Apache 2.0） |

---

### `/cmd` - 命令入口

| 文件 | 作用 |
|-----|-----|
| **emotionalBeach/** | 编译后的二进制可执行文件所在目录 |

---

### `/config` - 配置管理

| 文件 | 作用 |
|-----|-----|
| **config.go** | 配置结构体定义和加载逻辑（支持环境变量覆盖） |
| **config.yaml** | 配置文件（数据库、服务器端口、Redis、邮件等） |

**关键配置类**：
- `ServerConfig` - 服务器配置（端口、GitHub OAuth参数）
- `MailConfig` - 邮件配置（SMTP用户名、密码）
- `PostgresConfig` / `MySQLConfig` - 数据库配置
- `RedisConfig` - Redis 缓存配置

---

### `/docs` - API 文档

| 文件 | 作用 |
|-----|-----|
| **docs.go** | Swagger 文档生成配置 |
| **swagger.json** | Swagger 规范文件（JSON 格式） |
| **swagger.yaml** | Swagger 规范文件（YAML 格式） |

**说明**：通过 `swag` 工具自动从代码注解生成，访问 `/swagger/index.html` 查看接口文档。

---

### `/example` - 示例代码

| 文件 | 作用 |
|-----|-----|
| **client.go** | HTTP 客户端示例代码 |
| **totp.go** | TOTP（时间一次性密码）使用示例 |

---

### `/internal` - 核心业务代码

#### `/internal/common` - 工具函数

| 文件 | 作用 |
|-----|-----|
| **md5.go** | MD5 加密和盐值加密工具函数 |
| **valid_phone.go** | 手机号码格式验证函数 |

---

#### `/internal/controller` - 路由处理器（控制层）

| 文件 | 作用 |
|-----|-----|
| **user.go** | 用户相关接口处理（登录、注册、更新、删除、查询） |
| **github.go** | GitHub OAuth 登录和回调处理 |
| **relation.go** | 好友关系管理（获取好友列表、添加好友） |
| **preload_cache.go** | Redis 缓存预加载逻辑 |

---

#### `/internal/dao` - 数据访问层

| 文件 | 作用 |
|-----|-----|
| **user.go** | 用户表的 CRUD 操作（创建、查询、更新、删除） |
| **relation.go** | 好友关系表的数据库操作（查询、插入、删除） |

---

#### `/internal/global` - 全局资源

| 文件 | 作用 |
|-----|-----|
| **global.go** | 全局变量定义（Redis 客户端等） |
| **response.go** | 统一的响应格式工具函数（Success、Error） |

---

#### `/internal/initialize` - 初始化模块

| 文件 | 作用 |
|-----|-----|
| **logger.go** | 初始化 Zap 日志系统 |
| **manager.go** | 数据库、Redis 等资源管理 |
| **migration.go** | 数据库迁移（自动创建表） |

---

#### `/internal/middleware` - 中间件

| 文件 | 作用 |
|-----|-----|
| **jwt.go** | JWT Token 生成和验证中间件（7天有效期） |
| **cors.go** | CORS 跨域资源共享中间件 |
| **logger.go** | 请求日志记录中间件 |
| **assets_cache.go** | 静态资源缓存中间件 |
| **rateLimit.go** | IP 限流中间件（10秒内最多5次请求） |

---

#### `/internal/models` - 数据模型

| 文件 | 作用 |
|-----|-----|
| **user_basic.go** | 用户表和好友关系表的 GORM 模型定义 |

**关键模型**：
- `UserBasic` - 用户表（包含登录时间、心跳时间、登出时间等）
- `Relation` - 好友关系表（1=好友关系，2=群关系）
- `LoginRequest` - 登录请求结构体

---

#### `/internal/server` - 服务器配置

| 文件 | 作用 |
|-----|-----|
| **router.go** | 路由注册和分组（实现 RESTful API 路由树） |

**路由分组**：
- 无认证路由：`/ping`, `/`, `/login`, `/register`, `/login/github`, `/callback`
- v1 API 路由（需要 JWT 认证）：
  - `/v1/user/` - 用户管理
  - `/v1/relation/` - 好友关系
  - `/v1/api/webhook` - Webhook 接口

---

#### `/internal/service` - 业务逻辑层

| 文件 | 作用 |
|-----|-----|
| **webhook.go** | 邮件通知业务逻辑（支持多收件人、HTML内容解析） |

---

#### `/internal/templates` - 前端模板

| 文件 | 作用 |
|-----|-----|
| **index.html** | 主页 HTML 模板 |
| **templates.go** | 模板文件系统加载逻辑 |
| **assets/** | 前端资源文件 |
  - **cdnFallback.js** - CDN 资源加载备用方案脚本
  - **tailwindcss.js** - Tailwind CSS 配置

---

### `/tmp` - 临时文件

| 文件 | 作用 |
|-----|-----|
| **nginx.conf** | Nginx 反向代理配置（生产环境参考） |
| **scp_server** | SCP 服务器相关脚本 |

---

## 🔄 核心业务流程

### 用户注册流程
```
1. POST /register
   ├─ 验证用户名、密码、手机号
   ├─ MD5 + 盐值加密密码
   └─ 创建用户到数据库

2. 响应用户信息
```

### 用户登录流程
```
1. POST /login
   ├─ 查询用户
   ├─ 验证密码
   └─ 生成 JWT Token（7天有效期）

2. 响应 Token 和用户ID
```

### GitHub OAuth 登录流程
```
1. GET /login/github
   └─ 重定向到 GitHub 授权页

2. GitHub 回调 → GET /callback
   ├─ 交换授权码获取 Access Token
   ├─ 获取用户信息
   └─ 重定向到 Swagger 文档
```

### 好友关系管理
```
1. POST /v1/relation/list
   └─ 获取指定用户的好友列表

2. POST /v1/relation/add
   ├─ 按用户ID添加好友
   └─ 按用户名添加好友
```

### 邮件通知服务
```
1. POST /v1/api/webhook
   ├─ 验证收件人邮箱
   ├─ HTML 转纯文本
   └─ 通过 SMTP 发送邮件
```

---

## 🔒 安全特性

| 特性 | 实现 |
|-----|-----|
| **密码加密** | MD5 + 随机盐值 |
| **认证** | JWT Token（Bearer 方式） |
| **限流** | IP 限流中间件（10秒5次） |
| **CORS** | 跨域资源共享配置 |
| **日志** | Zap 结构化日志（含敏感信息过滤）|

---

## 📦 关键依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| gin-gonic/gin | v1.11.0 | HTTP 框架 |
| gorm.io/gorm | v1.31.1 | ORM 框架 |
| golang-jwt/jwt | v5.3.0 | JWT 令牌处理 |
| redis/go-redis | v9.17.2 | Redis 缓存 |
| go.uber.org/zap | v1.27.1 | 日志库 |
| spf13/viper | v1.21.0 | 配置管理 |
| swaggo/swag | v1.16.6 | Swagger 文档生成 |

---

## 🚀 快速启动

```bash
# 1. 生成 Swagger 文档
make gen

# 2. 构建 Docker 镜像
docker-compose up -d

# 或使用编译的二进制
make all
./cmd/emotionalBeach/emotionalBeach

# 3. 访问服务
# API: http://localhost:8080
# Swagger UI: http://localhost:8080/swagger/index.html
# 健康检查: http://localhost:8080/ping
```

---

## 💡 开发建议

1. **数据库迁移**：使用 `-migrate` 标志自动创建表
2. **热重载开发**：使用 `fresh` 工具进行热加载开发
3. **环境变量配置**：通过环境变量覆盖 config.yaml 中的配置
4. **API 文档**：始终在代码中维护 Swagger 注解，保持文档同步
5. **日志级别**：根据环境设置适当的日志级别

---

## 📝 许可证

Apache License 2.0

