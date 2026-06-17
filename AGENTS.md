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
```

Swagger generation must be re-run whenever Swag annotations (`// @Summary`, etc.) change:

```bash
swag init -o ./docs -g main.go
```

## Adding a New Feature (the required pattern)

1. **Model** (`internal/models/`) — add GORM struct; it is auto-migrated on `-migrate`.
2. **DAO** (`internal/dao/dao.go`) — add method to the `Dao` interface and implement in a new `dao/<domain>.go` file.
3. **Service** (`internal/service/`) — add method(s) to `*service.Service` in a new `service/<domain>.go`; embed any new
   config fields in `service.go:New()`.
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
users, err := svc.GetUsers()
...
}
```

## JWT & Route Protection

- Public routes registered directly on `r` (no middleware group).
- All `/v1/...` routes are guarded by `middleware.AuthJwt()` + IP rate-limiter (5 req / 10 s).
- JWT secret is injected from `cfg.Server.JWTSecret` via `middleware.SetJWTSecret()` in `server.New()`.
- `verifyToken` / `refreshToken` at `/auth/verify` and `/auth/refresh` handle their own token validation — **not**
  wrapped by `AuthJwt`.

## Configuration

All config lives in `config/config.yaml`. Environment variables override any field using the pattern `SECTION_FIELD` (
e.g., `SERVER_JWTSECRET`, `DATABASES_POSTGRES_MAIN_PASSWORD`). No `.env` file; no hardcoded secrets in code.

Redis can be disabled globally with `server.enableRedis: false`.

Multi-database support: add entries under `databases:` and set `default_database:`. Driver is chosen per-entry (
`type: postgres` or `type: mysql`).

## Key Files at a Glance

| File                          | Role                                                                |
|-------------------------------|---------------------------------------------------------------------|
| `internal/di/wire.go`         | Wire injector — source of truth for DI graph                        |
| `internal/di/app.go`          | Graceful shutdown: HTTP → DB → Redis → Logger                       |
| `internal/infra/startup.go`   | AutoMigrate, Redis cache preload, Prometheus collector registration |
| `internal/server/server.go`   | All route registrations in `initRouter()`                           |
| `internal/common/response.go` | The only 3 response helpers — use exclusively                       |
| `internal/middleware/jwt.go`  | Token generation, validation, refresh logic                         |
| `config/config.yaml`          | Single source of runtime configuration                              |

