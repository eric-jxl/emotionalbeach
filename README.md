# emotionalBeach

[📄 中文项目结构文档](./docs/PROJECT_STRUCTURE_CN.md) | [📄 English Project Structure](./docs/PROJECT_STRUCTURE_EN.md)

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/eric-jxl/emotionalbeach?color=blue&label=go&logo=go)
[![build-go-binary](https://github.com/eric-jxl/emotionalbeach/actions/workflows/go-binary-release.yml/badge.svg)](https://github.com/eric-jxl/emotionalbeach/actions/workflows/go-binary-release.yml)
[![Docker Image CI](https://github.com/eric-jxl/emotionalbeach/actions/workflows/docker-image.yml/badge.svg)](https://github.com/eric-jxl/emotionalbeach/actions/workflows/docker-image.yml)

## 📋 项目概览

**emotionalBeach** 是一个基于 Go 语言的后端 API 服务，使用 Gin 框架构建，集成了用户认证、好友关系管理、邮件通知等功能。该项目支持 GitHub OAuth 认证、JWT Token 验证、Redis 缓存、多数据库支持（PostgreSQL/MySQL）。

**开发环境：** Go v1.23.7 + Gin v1.10.1 + GORM v1.30.2 + Viper v1.20.1

### 🎯 核心功能

- ✅ **用户认证系统** - 支持用户名/密码登录和 GitHub OAuth 一键登录
- ✅ **好友关系管理** - 添加好友、获取好友列表等社交功能
- ✅ **JWT Token 认证** - 7天有效期的授权令牌
- ✅ **邮件通知服务** - 支持多收件人、HTML 内容的邮件发送
- ✅ **Redis 缓存** - 提高性能的缓存层
- ✅ **IP 限流保护** - 防止滥用和 DDoS 攻击
- ✅ **Swagger API 文档** - 自动生成的交互式 API 文档
- ✅ **多数据库支持** - 兼容 PostgreSQL 和 MySQL

### 📊 技术栈

| 技术 | 用途 |
|-----|-----|
| Gin v1.10.1 | HTTP 框架 |
| GORM v1.30.2 | ORM 数据库层 |
| JWT v5.3.0 | 令牌认证 |
| Redis v9.17.2 | 缓存层 |
| Viper v1.20.1 | 配置管理 |
| Zap v1.27.1 | 日志系统 |
| Swagger v1.16.6 | API 文档生成 |

---

## 🚀 快速启动

```bash
# 使用 Docker Compose (推荐)
docker compose up -d

# 或者拉取最新镜像
docker pull ghcr.io/eric-jxl/emotionalbeach:latest
```

> [!TIP]  
> - 新增 docker ci/cd 打包发布到 ghcr.io
> - 创建 release 自动编译跨平台二进制包
> - 启动前需要配置数据库，配置文件默认在 `config/config.yaml`
> - 可通过环境变量覆盖配置，默认 web 端口 8080

---

## 📚 项目大纲

### 目录结构

```
emotionalbeach/
├── main.go                  # 项目入口
├── go.mod / go.sum          # 依赖管理
├── config/                  # 配置管理
│   ├── config.go           # 配置加载
│   └── config.yaml         # 配置文件
├── internal/               # 核心业务代码
│   ├── controller/         # 路由控制器 (HTTP 请求处理)
│   ├── service/            # 业务逻辑层 (邮件服务等)
│   ├── dao/                # 数据访问层 (数据库操作)
│   ├── models/             # 数据模型定义
│   ├── middleware/         # 中间件 (JWT、限流、日志等)
│   ├── global/             # 全局变量和响应工具
│   ├── initialize/         # 初始化模块 (DB、Redis、日志)
│   └── templates/          # 前端模板和静态资源
├── docs/                   # API 文档 (自动生成)
├── example/                # 使用示例代码
└── tmp/                    # 临时配置文件
```

### 核心模块

| 模块 | 职责 | 文件 |
|-----|-----|-----|
| **Controller** | 处理 HTTP 请求，验证参数 | `user.go`, `github.go`, `relation.go` |
| **Service** | 业务逻辑处理 | `webhook.go` (邮件服务) |
| **DAO** | 数据库操作 | `user.go`, `relation.go` |
| **Middleware** | 请求拦截处理 | `jwt.go`, `rateLimit.go`, `cors.go` |
| **Models** | 数据结构定义 | `user_basic.go` (User、Relation) |

### API 接口概览

#### 👤 用户接口 (User API)

| 方法 | 端点 | 认证 | 功能 |
|-----|-----|-----|-----|
| POST | `/register` | ❌ | 用户注册 |
| POST | `/login` | ❌ | 用户登录 |
| GET | `/login/github` | ❌ | GitHub OAuth 登录 |
| GET | `/callback` | ❌ | GitHub 回调 |
| GET | `/v1/user/list` | ✅ | 获取所有用户 |
| GET | `/v1/user/condition` | ✅ | 条件查询用户 (ID/Email/Phone) |
| POST | `/v1/user/update` | ✅ | 更新用户信息 |
| DELETE | `/v1/user/delete` | ✅ | 删除用户 |

#### 👥 好友接口 (Relation API)

| 方法 | 端点 | 认证 | 功能 |
|-----|-----|-----|-----|
| POST | `/v1/relation/list` | ✅ | 获取好友列表 |
| POST | `/v1/relation/add` | ✅ | 添加好友 (ID/昵称) |

#### 📧 Webhook 接口

| 方法 | 端点 | 认证 | 功能 |
|-----|-----|-----|-----|
| POST | `/v1/api/webhook` | ✅ | 发送邮件通知 |

#### 🏥 其他接口

| 方法 | 端点 | 功能 |
|-----|-----|-----|
| GET | `/ping` | 健康检查 |
| GET | `/swagger/index.html` | Swagger API 文档 |

### 数据模型

#### UserBasic (用户表)
```
- ID: 用户ID (主键)
- Name: 用户名
- Password: 加密密码
- Phone: 手机号 (唯一、11位)
- Email: 邮箱
- Avatar: 头像 URL
- Gender: 性别 (male/female)
- Role: 角色 (user/admin/superadmin)
- LoginTime: 最后登录时间
- HeartBeatTime: 心跳时间
- LoginOutTime: 登出时间
- IsLogOut: 是否已登出
- DeviceInfo: 设备信息
```

#### Relation (好友关系表)
```
- ID: 关系ID (主键)
- OwnerId: 所有者ID
- TargetID: 目标用户ID
- Type: 关系类型 (1=好友, 2=群)
- Desc: 描述
```

---

## 🔐 安全特性

- 🔒 **密码加密**: MD5 + 随机盐值
- 🎫 **认证**: JWT Token (7天有效期)
- 🚫 **限流**: IP 限流 (10秒5次请求)
- 🌐 **CORS**: 跨域资源共享
- 📝 **日志**: Zap 结构化日志

---

## 📖 详细文档

更详细的项目结构说明和接口文档，请参考：

- [📄 中文项目结构文档](./docs/PROJECT_STRUCTURE_CN.md) - 完整的中文项目分析
- [📄 英文项目结构文档](./docs/PROJECT_STRUCTURE_EN.md) - 完整的英文项目分析
- [📋 中文 API 接口文档](./docs/API_DOCS_CN.md) - 详细的中文 API 说明和示例
- [📋 英文 API 接口文档](./docs/API_DOCS_EN.md) - Detailed English API documentation
- [📚 文档索引导航](./docs/INDEX.md) - 快速查找文档
- [📋 快速参考](./docs/QUICK_REFERENCE.md) - 常用命令和 API 速查

---

## 🛠️ 开发指南

### 1. 环境准备

```bash
# 安装依赖
go mod download

# 安装 Swagger 文档生成工具
go install github.com/swaggo/swag/cmd/swag@latest

# 安装热重载工具 (可选，用于开发环境)
go install github.com/zzwx/fresh@latest
```

### 2. 数据库配置

编辑 `config/config.yaml`，配置数据库连接信息：

```yaml
database:
  type: postgres  # 或 mysql
  host: localhost
  port: 5432
  user: postgres
  password: password
  dbname: emotionalbeach
```

### 3. 数据库迁移

```bash
# 首次运行需要创建数据库表
go run main.go -migrate
```

### 4. 运行服务

```bash
# 方式1: 直接运行
go run main.go

# 方式2: 使用 Docker Compose (推荐)
docker-compose up -d

# 方式3: 编译后运行
make all
./cmd/emotionalBeach/emotionalBeach

# 方式4: 热重载开发模式
fresh -generate  # 首次生成配置文件
fresh -c .fresh.yaml
```

### 5. 生成 Swagger 文档

```bash
# 使用 Makefile
make gen

# 或直接使用 swag 命令
swag init -o ./docs -g main.go
```

### 6. 常用编译命令

```bash
make all  # 打包编译并 upx 压缩
make gen  # 生成 Swagger 文档
```

### 7. Git 仓库设置

```bash
cd existing_repo
git remote add origin https://github.com/eric-jxl/emotionalbeach.git
git branch -M main
git push -uf origin main
```

---

## 📞 贡献指南

欢迎提交 Issue 和 Pull Request！

---

## 📄 许可证

Apache License 2.0
