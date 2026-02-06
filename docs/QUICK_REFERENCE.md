# 📚 emotionalBeach 文档快速参考

## 🎯 文档一览表

| 文档 | 用途 | 阅读时间 | 推荐人群 |
|------|------|---------|---------|
| **README.md** | 项目总览、快速启动 | 10分钟 | 所有人 |
| **INDEX.md** | 文档导航、快速查询 | 5分钟 | 第一次使用 |
| **PROJECT_STRUCTURE_CN.md** | 代码结构分析 | 30分钟 | 开发者 |
| **PROJECT_STRUCTURE_EN.md** | 代码结构分析 (英文) | 30分钟 | 国际开发者 |
| **API_DOCS_CN.md** | API 详细文档 | 1小时 | API 使用者 |
| **API_DOCS_EN.md** | API 详细文档 (英文) | 1小时 | 国际 API 使用者 |
| **DOCUMENTATION_SUMMARY.md** | 文档总结、最佳实践 | 20分钟 | 项目管理者 |

---

## ⚡ 快速命令

```bash
# 生成 Swagger 文档
make gen

# 启动服务
docker-compose up -d

# 直接运行
go run main.go

# 数据库迁移
go run main.go -migrate

# 编译二进制
make all
```

---

## 🔌 API 快速查询

### 认证接口
- `POST /register` - 注册
- `POST /login` - 登录
- `GET /login/github` - GitHub OAuth

### 用户接口 (需认证)
- `GET /v1/user/list` - 列表
- `GET /v1/user/condition` - 查询
- `POST /v1/user/update` - 更新
- `DELETE /v1/user/delete` - 删除

### 好友接口 (需认证)
- `POST /v1/relation/list` - 列表
- `POST /v1/relation/add` - 添加

### 其他接口
- `POST /v1/api/webhook` - 邮件 (需认证)
- `GET /ping` - 健康检查
- `GET /swagger/index.html` - Swagger UI

---

## 📋 常用配置

```yaml
# config/config.yaml

server:
  port: 8080
  enableRedis: true

database:
  type: postgres
  host: localhost
  port: 5432
  user: postgres
  password: password
  dbname: emotionalbeach

redis:
  host: localhost
  port: 6379
  password: ""
```

---

## 🔐 安全相关

### 密码加密
- 算法: MD5 + 随机盐值
- 盐值保存: 用户表

### Token 认证
- 类型: JWT (HS256)
- 有效期: 7 天
- 使用方式: `Authorization: <token>`

### 限流保护
- 方式: IP 限流
- 规则: 10秒内最多5次请求
- 超限状态码: 429

---

## 🚀 快速启动步骤

```bash
# 1. 配置数据库
编辑 config/config.yaml

# 2. 启动服务
docker-compose up -d

# 3. 数据库迁移 (可选)
go run main.go -migrate

# 4. 访问应用
浏览器: http://localhost:8080
Swagger: http://localhost:8080/swagger/index.html

# 5. 健康检查
curl http://localhost:8080/ping
```

---

## 📊 项目统计

- 📦 依赖: 20+ 个
- 📁 模块: 15+ 个
- 🔌 API: 14 个
- 📝 文档: 8 个
- 📄 代码: 3000+ 行
- 📚 文档: 3600+ 行

---

## 🎓 学习路径

### 初级开发者
1. 阅读 README.md
2. 运行 make gen
3. 访问 Swagger UI
4. 测试一个 API

### 中级开发者
1. 阅读 PROJECT_STRUCTURE_CN.md
2. 查看源代码结构
3. 理解 MVC 分层
4. 实现简单功能

### 高级开发者
1. 深入阅读代码
2. 理解中间件设计
3. 优化性能
4. 贡献代码

---

## 🛠️ 常见问题速查

| 问题 | 答案 | 文档位置 |
|------|------|---------|
| 如何注册用户？ | POST /register | API_DOCS_CN.md 1.1 |
| 如何登录？ | POST /login | API_DOCS_CN.md 1.2 |
| 如何获取 Token？ | 登录时返回 | API_DOCS_CN.md 1.2 |
| 如何调用 API？ | 使用 curl 或 HTTP 客户端 | API_DOCS_CN.md 示例 |
| 如何部署？ | docker-compose up | README.md 快速启动 |
| 如何配置数据库？ | 编辑 config.yaml | README.md 开发指南 |

---

## 📞 获取帮助

### 快速导航
👉 打开 [INDEX.md](./INDEX.md)

### API 文档
👉 打开 [API_DOCS_CN.md](./API_DOCS_CN.md)

### 项目结构
👉 打开 [PROJECT_STRUCTURE_CN.md](./PROJECT_STRUCTURE_CN.md)

### Swagger UI
👉 访问 http://localhost:8080/swagger/index.html

---

## 🌍 语言选择

- 🇨🇳 中文: PROJECT_STRUCTURE_CN.md, API_DOCS_CN.md
- 🇬🇧 英文: PROJECT_STRUCTURE_EN.md, API_DOCS_EN.md

---

**最后更新**: 2024年2月6日  
**版本**: 1.0  
**生成工具**: GitHub Copilot

---

👉 **下一步**: 打开 [README.md](../README.md) 开始！

