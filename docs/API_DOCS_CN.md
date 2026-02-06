# emotionalBeach API 接口文档 - 中文版

## 📌 文档说明

- **项目名称**: emotionalBeach
- **API 版本**: v1.0
- **基础 URL**: `http://localhost:8080`
- **认证方式**: JWT Token (Bearer Token)
- **默认端口**: 8080

---

## 🔑 认证信息

### JWT Token 获取

1. 用户注册或登录后获得 Token
2. 在所有需要认证的请求头中添加：`Authorization: <token>`

### Token 有效期

- **有效期**: 7 天
- **过期后**: 需要重新登录获取新 Token

---

## 📍 接口清单

### 一、用户认证接口

#### 1.1 用户注册

**端点**: `POST /register`

**认证**: ❌ 不需要

**请求方式**: `multipart/form-data`

**请求参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| name | string | ✅ | 用户名 |
| password | string | ✅ | 密码 |
| repeat_password | string | ✅ | 确认密码（需与密码一致） |
| phone | string | ✅ | 手机号（11位） |
| email | string | ❌ | 邮箱 |

**请求示例**:

```bash
curl -X POST http://localhost:8080/register \
  -F "name=张三" \
  -F "password=123456" \
  -F "repeat_password=123456" \
  -F "phone=13800138000" \
  -F "email=user@example.com"
```

**响应示例**:

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "name": "张三",
    "phone": "13800138000",
    "email": "user@example.com",
    "avatar": "",
    "gender": "male",
    "role": "user"
  },
  "message": "新增用户成功！"
}
```

**错误处理**:

| 错误码 | 说明 |
|--------|------|
| 401 | 用户名或密码或确认密码不能为空 |
| 401 | 两次密码不一致 |
| 401 | 手机号不能为空 |
| 401 | 手机号必须为11位 |
| 403 | 手机号非法 |
| 401 | 该用户已注册 |

---

#### 1.2 用户登录

**端点**: `POST /login`

**认证**: ❌ 不需要

**请求方式**: `application/json`

**请求参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| username | string | ✅ | 用户名 |
| password | string | ✅ | 密码 |

**请求示例**:

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "张三",
    "password": "123456"
  }'
```

**响应示例**:

```json
{
  "code": 200,
  "message": "登录成功",
  "user_id": 1,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**错误处理**:

| 错误码 | 说明 |
|--------|------|
| 401 | Invalid request 请求格式错误 |
| 403 | 登录失败 |
| 404 | 用户名不存在 |
| 401 | 密码错误 |

---

#### 1.3 GitHub OAuth 登录

**端点**: `GET /login/github`

**认证**: ❌ 不需要

**功能**: 重定向到 GitHub 授权页面

**请求示例**:

```bash
curl -X GET http://localhost:8080/login/github
```

**说明**:
- 用户会被重定向到 GitHub OAuth 授权页
- 授权后会回调 `/callback` 端点
- 最终重定向到 Swagger 文档页面

---

#### 1.4 GitHub 回调接口

**端点**: `GET /callback`

**认证**: ❌ 不需要

**查询参数**:

| 参数 | 说明 |
|------|------|
| code | GitHub 授权返回的 code |

**功能**: 处理 GitHub OAuth 回调，获取用户信息并生成 Token

**说明**: 
- 此接口由 GitHub 自动调用
- 成功后重定向到 Swagger 文档

---

### 二、用户管理接口

#### 2.1 获取所有用户

**端点**: `GET /v1/user/list`

**认证**: ✅ 需要 JWT Token

**请求头**:

```
Authorization: <token>
```

**请求示例**:

```bash
curl -X GET http://localhost:8080/v1/user/list \
  -H "Authorization: your_token_here"
```

**响应示例**:

```json
{
  "code": 200,
  "data": [
    {
      "name": "张三",
      "avatar": "",
      "gender": "male",
      "phone": "13800138000",
      "email": "user@example.com",
      "identity": ""
    },
    {
      "name": "李四",
      "avatar": "",
      "gender": "female",
      "phone": "13900139000",
      "email": "user2@example.com",
      "identity": ""
    }
  ],
  "message": "success"
}
```

---

#### 2.2 条件查询用户

**端点**: `GET /v1/user/condition`

**认证**: ✅ 需要 JWT Token

**查询参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| id | string | ❌ | 用户 ID |
| email | string | ❌ | 邮箱 |
| phone | string | ❌ | 手机号 |

**注意**: 至少需要提供一个查询参数

**请求示例**:

```bash
# 按 ID 查询
curl -X GET "http://localhost:8080/v1/user/condition?id=1" \
  -H "Authorization: your_token_here"

# 按邮箱查询
curl -X GET "http://localhost:8080/v1/user/condition?email=user@example.com" \
  -H "Authorization: your_token_here"

# 按手机号查询
curl -X GET "http://localhost:8080/v1/user/condition?phone=13800138000" \
  -H "Authorization: your_token_here"
```

**响应示例**:

```json
{
  "code": 200,
  "data": {
    "name": "张三",
    "avatar": "",
    "gender": "male",
    "phone": "13800138000",
    "email": "user@example.com",
    "identity": ""
  },
  "message": "success"
}
```

**错误处理**:

| 错误码 | 说明 |
|--------|------|
| 400 | 至少需要一个参数（id、电子邮件或电话） |
| 500 | 查询失败 |

---

#### 2.3 更新用户信息

**端点**: `POST /v1/user/update`

**认证**: ✅ 需要 JWT Token

**请求方式**: `multipart/form-data`

**请求参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| id | string | ✅ | 用户 ID |
| name | string | ❌ | 用户名 |
| password | string | ❌ | 新密码 |
| phone | string | ❌ | 手机号 |
| email | string | ❌ | 邮箱 |
| avatar | string | ❌ | 头像 URL |
| gender | string | ❌ | 性别 (male/female) |

**请求示例**:

```bash
curl -X POST http://localhost:8080/v1/user/update \
  -H "Authorization: your_token_here" \
  -F "id=1" \
  -F "name=张三新名字" \
  -F "avatar=https://example.com/avatar.jpg" \
  -F "gender=male"
```

**响应示例**:

```json
{
  "code": 200,
  "data": {
    "uid": 1
  },
  "message": "success"
}
```

**错误处理**:

| 错误码 | 说明 |
|--------|------|
| 500 | 修改信息失败 |

---

#### 2.4 删除用户

**端点**: `DELETE /v1/user/delete`

**认证**: ✅ 需要 JWT Token

**查询参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| id | uint | ✅ | 用户 ID |

**请求示例**:

```bash
curl -X DELETE "http://localhost:8080/v1/user/delete?id=1" \
  -H "Authorization: your_token_here"
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success"
}
```

---

### 三、好友关系接口

#### 3.1 获取好友列表

**端点**: `POST /v1/relation/list`

**认证**: ✅ 需要 JWT Token

**请求方式**: `multipart/form-data`

**请求参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| userId | uint | ✅ | 用户 ID |

**请求示例**:

```bash
curl -X POST http://localhost:8080/v1/relation/list \
  -H "Authorization: your_token_here" \
  -F "userId=1"
```

**响应示例**:

```json
{
  "code": 200,
  "data": {
    "count": 2,
    "users": [
      {
        "name": "李四",
        "avatar": "",
        "gender": "female",
        "phone": "13900139000",
        "email": "user2@example.com",
        "identity": ""
      },
      {
        "name": "王五",
        "avatar": "",
        "gender": "male",
        "phone": "14000140000",
        "email": "user3@example.com",
        "identity": ""
      }
    ]
  },
  "message": "success"
}
```

**错误处理**:

| 错误码 | 说明 |
|--------|------|
| 404 | 好友为空 |

---

#### 3.2 添加好友

**端点**: `POST /v1/relation/add`

**认证**: ✅ 需要 JWT Token

**请求方式**: `multipart/form-data`

**请求参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| userId | uint | ✅ | 当前用户 ID |
| targetName | string/uint | ✅ | 目标用户 ID 或用户名 |

**请求示例**:

```bash
# 按用户 ID 添加
curl -X POST http://localhost:8080/v1/relation/add \
  -H "Authorization: your_token_here" \
  -F "userId=1" \
  -F "targetName=2"

# 按用户名添加
curl -X POST http://localhost:8080/v1/relation/add \
  -H "Authorization: your_token_here" \
  -F "userId=1" \
  -F "targetName=李四"
```

**响应示例**:

```json
{
  "code": 200,
  "data": {
    "msg": "添加好友成功"
  },
  "message": "success"
}
```

---

### 四、Webhook 邮件接口

#### 4.1 发送邮件通知

**端点**: `POST /v1/api/webhook`

**认证**: ✅ 需要 JWT Token

**请求方式**: `application/json`

**请求参数**:

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| title | string | ✅ | 邮件标题 |
| content | string | ❌ | 邮件内容（支持 HTML） |
| receivers | string[] | ✅ | 收件人邮箱列表 |

**请求示例**:

```bash
curl -X POST http://localhost:8080/v1/api/webhook \
  -H "Authorization: your_token_here" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "欢迎使用 emotionalBeach",
    "content": "<h1>欢迎！</h1><p>这是一封 HTML 邮件</p>",
    "receivers": [
      "user@example.com",
      "another@example.com"
    ]
  }'
```

**响应示例**:

```json
{
  "code": 200,
  "data": {
    "message": "邮件发送成功"
  },
  "message": "success"
}
```

**邮件功能说明**:
- ✅ 支持多收件人
- ✅ 支持 HTML 内容
- ✅ 自动 HTML 转纯文本
- ✅ 邮箱格式校验
- ✅ 异步发送（同步等待）

**错误处理**:

| 错误码 | 说明 |
|--------|------|
| 400 | 邮件标题不能为空 |
| 400 | 收件人列表为空 |
| 400 | 无效的邮箱地址 |
| 500 | 邮件发送失败 |

---

### 五、其他接口

#### 5.1 健康检查

**端点**: `GET /ping`

**认证**: ❌ 不需要

**功能**: 检查服务器是否运行

**请求示例**:

```bash
curl -X GET http://localhost:8080/ping
```

**响应示例**:

```json
{
  "message": "pong"
}
```

---

#### 5.2 获取主页

**端点**: `GET /`

**认证**: ❌ 不需要

**功能**: 返回前端主页 HTML

**请求示例**:

```bash
curl -X GET http://localhost:8080/
```

---

#### 5.3 Swagger API 文档

**端点**: `GET /swagger/index.html`

**认证**: ❌ 不需要

**功能**: 交互式 API 文档

**访问方式**: 在浏览器中打开 `http://localhost:8080/swagger/index.html`

---

## 📊 响应格式规范

### 成功响应

```json
{
  "code": 200,
  "data": {},  // 响应数据
  "message": "success"
}
```

### 错误响应

```json
{
  "code": 400,  // 或其他错误码
  "message": "错误说明"
}
```

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（Token 缺失或无效） |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁（限流） |
| 500 | 服务器内部错误 |

---

## 🔒 安全性说明

### 密码安全

- ✅ 密码采用 MD5 + 随机盐值加密
- ✅ 盐值在用户表中单独存储
- ✅ 不支持密码找回（请妥善保管密码）

### Token 安全

- ✅ JWT Token 使用 HS256 算法签名
- ✅ 有效期为 7 天
- ✅ 过期 Token 需重新登录获取

### 请求限制

- ✅ IP 限流：10 秒内最多 5 次请求
- ✅ 超出限制返回 429 状态码
- ✅ 可根据 IP 自动恢复

### 数据验证

- ✅ 手机号：必须为 11 位，格式检验
- ✅ 邮箱：正则表达式格式检验
- ✅ 用户名：不允许为空
- ✅ 密码：不允许为空

---

## 🧪 测试流程

### 1. 注册测试

```bash
# 注册用户
curl -X POST http://localhost:8080/register \
  -F "name=testuser" \
  -F "password=123456" \
  -F "repeat_password=123456" \
  -F "phone=13800138000" \
  -F "email=test@example.com"
```

### 2. 登录测试

```bash
# 登录获取 Token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"123456"}'
```

### 3. 获取用户列表

```bash
# 使用 Token 调用受保护接口
curl -X GET http://localhost:8080/v1/user/list \
  -H "Authorization: <your_token>"
```

### 4. 添加好友

```bash
# 需要至少有两个用户
curl -X POST http://localhost:8080/v1/relation/add \
  -H "Authorization: <your_token>" \
  -F "userId=1" \
  -F "targetName=2"
```

### 5. 发送邮件

```bash
# 发送邮件通知
curl -X POST http://localhost:8080/v1/api/webhook \
  -H "Authorization: <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"测试邮件",
    "content":"这是一封测试邮件",
    "receivers":["test@example.com"]
  }'
```

---

## 📞 常见问题

### Q1: Token 过期怎么办？

A: Token 有 7 天有效期，过期后需要重新登录。建议在客户端实现 Token 刷新逻辑。

### Q2: 忘记密码怎么办？

A: 目前不支持密码找回功能，请确保密码妥善保管。

### Q3: 如何添加好友？

A: 使用 `/v1/relation/add` 接口，可以按用户 ID 或用户名添加。

### Q4: 发送邮件失败怎么办？

A: 检查邮箱地址格式是否正确，确保 SMTP 配置无误。

### Q5: 请求频繁返回 429？

A: 您的 IP 请求超出限流阈值（10 秒 5 次），请稍候重试。

---

## 📝 更新日志

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0 | 2024-01-01 | 初始版本发布 |

---

## 📄 相关资源

- [项目结构文档](./PROJECT_STRUCTURE_CN.md)
- [GitHub Repository](https://github.com/eric-jxl/emotionalbeach)
- [Swagger UI](http://localhost:8080/swagger/index.html)

---

**文档最后更新**: 2024年

