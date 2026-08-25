# AGENTS.md — emotionalBeach Codebase Guide

## Architecture Overview

Five-layer dependency graph, assembled at compile time via **Google Wire**:

```
main.go → di.InitializeApp(cfg)
  ├── infra.Provider   → *gorm.DB, *redis.Client, *infra.Loggers
  ├── dao.Provider     → dao.Dao  (interface wrapping GORM queries)
  ├── service.Provider → *service.Service  (all business logic, single struct)
  ├── server.New       → *http.Server  (Gin engine + all routes)
  └── di.NewApp        → *App  (graceful shutdown orchestrator)
```

**Never edit `internal/di/wire_gen.go` by hand** — it is auto-generated. After changing providers, run:

```bash
wire gen ./internal/di/...
```

## Critical Developer Commands

```bash
go run main.go              # run locally (reads config/config.yaml)
go run main.go -migrate     # run DB AutoMigrate then exit
make all                    # compile + upx-compress → cmd/emotionalBeach
make gen                    # swag init → regenerates docs/
make fmt                    # gofmt across all packages
docker compose up -d        # full stack: app + PostgreSQL + Prometheus + Grafana
# Kubernetes (Kustomize) — manifests in k8s/, see k8s/README.md
make k8s-apply              # kubectl apply -k → apply all k8s manifests
make k8s-deploy TAG=v1.0.0  # build+push image, rolling update, auto-rollback on failure
make k8s-status             # rollout status
make k8s-rollback           # undo to previous revision
```

Swagger generation must be re-run whenever Swag annotations (`// @Summary`, etc.) change:

```bash
swag init -o ./docs -g main.go
```

## Adding a New Feature (the required pattern)

1. **Model** (`internal/models/`) — add GORM struct; it is auto-migrated on `-migrate`.
   Embed `BaseModel` (ID) and `DatetimeModel` (CreatedAt/UpdatedAt/DeletedAt soft-delete) from
   `internal/models/user_basic.go`. **Register the new struct in the `AutoMigrate` call in
   `internal/infra/startup.go`** — otherwise the table is never created.
2. **DAO** (`internal/dao/dao.go`) — add method to the `Dao` interface and implement in a new `dao/<domain>.go` file.
3. **Service** (`internal/service/`) — add method(s) to `*service.Service` in a new `service/<domain>.go`; embed any new
   config fields in `service.go:New()`. Add Prometheus counters to `internal/infra/metrics.go` (namespace `eb`) and
   increment them labelled by outcome. For passwords use `common.Sha256Password()`; legacy MD5+salt hashes are
   auto-migrated to sha256 on successful login (`service.Login`).
4. **Handler** (`internal/server/`) — add handler functions (not methods; the package-level `var svc *service.Service`
   is the shared dependency) and register routes in `server.go:initRouter()`.
5. **No new Wire providers needed** unless adding infrastructure (DB, Redis, etc.).

## Unified API Response — mandatory

All handlers **must** use exactly these three helpers from `internal/common/response.go`:

| Helper                            | When         | HTTP status | `code` field        |
|-----------------------------------|--------------|-------------|---------------------|
| `common.Success(c, data)`         | success      | 200         | `200`               |
| `common.Fail(c, httpStatus, msg)` | client error | 4xx         | same as HTTP status |
| `common.ServerError(c, msg)`      | server error | 500         | `500`               |

Do **not** call `c.JSON` directly in handlers.

## Handler Pattern (no structs)

Handlers are **package-level functions**, not methods on a struct. The package-level `svc` variable is assigned once in
`server.New()`:

```go
// internal/server/server.go
var svc *service.Service // shared by all handlers in this package

func getUsers(c *gin.Context) {
list, err := svc.GetList()
if err != nil {
common.ServerError(c, err.Error())
return
}
// ...
}
```

## JWT & Route Protection

- Global middleware chain (applied to every route in `server.New()`): `RequestID()` → `gin.Recovery()` →
  `PanicRecoveryMiddleware()` → `PrometheusMiddleware()` → `ZapLogger()`.
- Routes are registered in `initRouter()` via four registrars by concern:
  `registerSystemRoutes` (`/ping`, `/health`, `/metrics`), `registerStaticRoutes` (`/`, `/assets/*`, `/swagger/*`),
  `registerPublicAuthRoutes` (`/login`, `/register`, `/login/github`, `/callback`, `/auth/*`), and
  `registerProtectedV1Routes` (all `/v1/...`).
- All `/v1/...` routes are guarded by `middleware.AuthJwt()` + IP rate-limiter. Rate-limit policy is config-driven
  (`server.rateLimitWindowSec` / `server.rateLimitMaxReq`; default 10 s / 5 req).
- JWT secret is injected from `cfg.Server.JWTSecret` via `middleware.SetJWTSecret()` in `server.New()`.
- `verifyToken` / `refreshToken` at `/auth/verify` and `/auth/refresh` handle their own token validation — **not**
  wrapped by `AuthJwt`.
- `middleware.ParseToken` caches validated claims in memory (keyed by raw token); `SetJWTSecret` flushes the cache on
  secret rotation.

## Configuration

All config lives in `config/config.yaml`. Environment variables override any field using the pattern `SECTION_FIELD` (
e.g., `SERVER_JWTSECRET`, `DATABASES_POSTGRES_MAIN_PASSWORD`). No `.env` file; no hardcoded secrets in code.

Key sections: `server` (port, `jwtSecret`, GitHub OAuth `clientID`/`clientSecret`, ESA captcha `esaCaptcha*`,
rate-limit, HTTP timeouts), `log` (level + lumberjack rotating-file), `databases`, `redis`, `mail` (SMTP credentials
for the webhook email service).

Redis can be disabled globally with `server.enableRedis: false`.

Multi-database support: add entries under `databases:` and set `default_database:`. Driver is chosen per-entry (
`type: postgres` or `type: mysql`).

Sensible defaults for timeouts, rate limits, and log rotation are defined in `config.go:defaultValues` and applied at
load time.

## Key Files at a Glance

| File                                | Role                                                                |
|-------------------------------------|---------------------------------------------------------------------|
| `internal/di/wire.go`               | Wire injector — source of truth for DI graph                        |
| `internal/di/app.go`                | Graceful shutdown: HTTP → DB → Redis → Logger                       |
| `internal/infra/startup.go`         | AutoMigrate, Redis cache preload, Prometheus collector registration |
| `internal/infra/metrics.go`         | Prometheus counters/gauges (namespace `eb`) + DB-pool collector     |
| `internal/server/server.go`         | All route registrations in `initRouter()`                           |
| `internal/common/response.go`       | The only 3 response helpers — use exclusively                       |
| `internal/common/md5.go`            | sha256 + legacy MD5 password hashing, phone validation              |
| `internal/middleware/jwt.go`        | Token generation, validation, refresh, in-memory parse cache        |
| `internal/middleware/rateLimit.go`  | IP token-bucket rate limiter                                        |
| `internal/middleware/request_id.go` | Per-request `X-Request-Id` generation/propagation                   |
| `internal/models/user_basic.go`     | GORM models: `BaseModel`, `DatetimeModel`, `UserBasic`, `Relation`  |
| `internal/templates/templates.go`   | `embed.FS` for login page, Swagger UI, and static assets            |
| `config/config.yaml`                | Single source of runtime configuration                              |
