# emotionalBeach - Project Structure Documentation

## 📋 Project Overview

**emotionalBeach** is a Go-based backend API service built with the Gin framework, featuring user authentication, friend relationship management, and email notification services. The project supports GitHub OAuth authentication, JWT token verification, Redis caching, and multiple database support (PostgreSQL/MySQL).

**Development Environment:** Go v1.23.7 + Gin v1.10.1 + GORM v1.30.2 + Viper v1.20.1

---

## 📁 Directory Structure Explanation

### Root Directory Files

| File/Directory | Purpose |
|----------------|---------|
| **main.go** | Project entry point, initializes logging, database, and Redis, starts HTTP server |
| **go.mod** | Go module dependency declaration file |
| **go.sum** | Go module dependency version lock file |
| **Makefile** | Compilation shortcuts (e.g., `make all`, `make gen`) |
| **Dockerfile** | Docker image build file |
| **docker-compose.yml** | Docker Compose configuration for local development |
| **deploy.sh** | Deployment script |
| **entrypoint.sh** | Docker container startup script |
| **README.md** | Project documentation |
| **LICENSE** | Open source license (Apache 2.0) |

---

### `/cmd` - Command Entry

| File | Purpose |
|------|---------|
| **emotionalBeach/** | Directory containing compiled executable binary |

---

### `/config` - Configuration Management

| File | Purpose |
|------|---------|
| **config.go** | Configuration struct definitions and loading logic (supports environment variable overrides) |
| **config.yaml** | Configuration file (database, server port, Redis, email, etc.) |

**Key Configuration Classes:**
- `ServerConfig` - Server configuration (port, GitHub OAuth parameters)
- `MailConfig` - Email configuration (SMTP username, password)
- `PostgresConfig` / `MySQLConfig` - Database configurations
- `RedisConfig` - Redis cache configuration

---

### `/docs` - API Documentation

| File | Purpose |
|------|---------|
| **docs.go** | Swagger documentation generation configuration |
| **swagger.json** | Swagger specification (JSON format) |
| **swagger.yaml** | Swagger specification (YAML format) |

**Note:** Auto-generated from code annotations using the `swag` tool. Access `/swagger/index.html` to view API documentation.

---

### `/example` - Example Code

| File | Purpose |
|------|---------|
| **client.go** | HTTP client example code |
| **totp.go** | TOTP (Time-based One-Time Password) usage example |

---

### `/internal` - Core Business Logic

#### `/internal/common` - Utility Functions

| File | Purpose |
|------|---------|
| **md5.go** | MD5 encryption and salt encryption utility functions |
| **valid_phone.go** | Phone number format validation function |

---

#### `/internal/controller` - Route Handlers (Controller Layer)

| File | Purpose |
|------|---------|
| **user.go** | User-related endpoint handlers (login, registration, update, delete, query) |
| **github.go** | GitHub OAuth login and callback handling |
| **relation.go** | Friend relationship management (get friend list, add friend) |
| **preload_cache.go** | Redis cache preloading logic |

---

#### `/internal/dao` - Data Access Layer

| File | Purpose |
|------|---------|
| **user.go** | User table CRUD operations (create, read, update, delete) |
| **relation.go** | Friend relationship table database operations (query, insert, delete) |

---

#### `/internal/global` - Global Resources

| File | Purpose |
|------|---------|
| **global.go** | Global variable definitions (Redis client, etc.) |
| **response.go** | Unified response format utility functions (Success, Error) |

---

#### `/internal/initialize` - Initialization Module

| File | Purpose |
|------|---------|
| **logger.go** | Zap logging system initialization |
| **manager.go** | Database and Redis resource management |
| **migration.go** | Database migration (auto-create tables) |

---

#### `/internal/middleware` - Middleware

| File | Purpose |
|------|---------|
| **jwt.go** | JWT token generation and verification middleware (7-day validity) |
| **cors.go** | CORS (Cross-Origin Resource Sharing) middleware |
| **logger.go** | Request logging middleware |
| **assets_cache.go** | Static asset caching middleware |
| **rateLimit.go** | IP rate limiting middleware (max 5 requests per 10 seconds) |

---

#### `/internal/models` - Data Models

| File | Purpose |
|------|---------|
| **user_basic.go** | GORM model definitions for user and friend relationship tables |

**Key Models:**
- `UserBasic` - User table (includes login time, heartbeat time, logout time, etc.)
- `Relation` - Friend relationship table (1=friend relationship, 2=group relationship)
- `LoginRequest` - Login request structure

---

#### `/internal/server` - Server Configuration

| File | Purpose |
|------|---------|
| **router.go** | Route registration and grouping (implements RESTful API route tree) |

**Route Groups:**
- Unauthenticated routes: `/ping`, `/`, `/login`, `/register`, `/login/github`, `/callback`
- v1 API routes (require JWT authentication):
  - `/v1/user/` - User management
  - `/v1/relation/` - Friend relationships
  - `/v1/api/webhook` - Webhook endpoints

---

#### `/internal/service` - Business Logic Layer

| File | Purpose |
|------|---------|
| **webhook.go** | Email notification business logic (supports multiple recipients, HTML content parsing) |

---

#### `/internal/templates` - Frontend Templates

| File | Purpose |
|------|---------|
| **index.html** | Main page HTML template |
| **templates.go** | Template file system loading logic |
| **assets/** | Frontend resource files |
  - **cdnFallback.js** - CDN resource fallback loading script
  - **tailwindcss.js** - Tailwind CSS configuration

---

### `/tmp` - Temporary Files

| File | Purpose |
|------|---------|
| **nginx.conf** | Nginx reverse proxy configuration (production reference) |
| **scp_server** | SCP server-related scripts |

---

## 🔄 Core Business Flow

### User Registration Flow
```
1. POST /register
   ├─ Validate username, password, phone number
   ├─ Encrypt password with MD5 + salt
   └─ Create user in database

2. Return user information
```

### User Login Flow
```
1. POST /login
   ├─ Query user
   ├─ Verify password
   └─ Generate JWT Token (7-day validity)

2. Return Token and user ID
```

### GitHub OAuth Login Flow
```
1. GET /login/github
   └─ Redirect to GitHub authorization page

2. GitHub callback → GET /callback
   ├─ Exchange authorization code for Access Token
   ├─ Fetch user information
   └─ Redirect to Swagger documentation
```

### Friend Relationship Management
```
1. POST /v1/relation/list
   └─ Get friend list for specified user

2. POST /v1/relation/add
   ├─ Add friend by user ID
   └─ Add friend by username
```

### Email Notification Service
```
1. POST /v1/api/webhook
   ├─ Validate recipient email addresses
   ├─ Convert HTML to plain text
   └─ Send email via SMTP
```

---

## 🔒 Security Features

| Feature | Implementation |
|---------|-----------------|
| **Password Encryption** | MD5 + random salt |
| **Authentication** | JWT Token (Bearer method) |
| **Rate Limiting** | IP rate limiting middleware (5 requests per 10 seconds) |
| **CORS** | Cross-origin resource sharing configuration |
| **Logging** | Zap structured logging (with sensitive data filtering) |

---

## 📦 Key Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| gin-gonic/gin | v1.11.0 | HTTP framework |
| gorm.io/gorm | v1.31.1 | ORM framework |
| golang-jwt/jwt | v5.3.0 | JWT token handling |
| redis/go-redis | v9.17.2 | Redis caching |
| go.uber.org/zap | v1.27.1 | Logging library |
| spf13/viper | v1.21.0 | Configuration management |
| swaggo/swag | v1.16.6 | Swagger documentation generation |

---

## 🚀 Quick Start

```bash
# 1. Generate Swagger documentation
make gen

# 2. Build Docker image
docker-compose up -d

# Or use compiled binary
make all
./cmd/emotionalBeach/emotionalBeach

# 3. Access the service
# API: http://localhost:8080
# Swagger UI: http://localhost:8080/swagger/index.html
# Health check: http://localhost:8080/ping
```

---

## 💡 Development Tips

1. **Database Migration**: Use the `-migrate` flag to auto-create tables
2. **Hot Reload Development**: Use the `fresh` tool for hot reload during development
3. **Environment Variables**: Override config.yaml settings via environment variables
4. **API Documentation**: Always maintain Swagger annotations in code to keep documentation in sync
5. **Log Level**: Set appropriate log levels based on environment

---

## 📝 License

Apache License 2.0

