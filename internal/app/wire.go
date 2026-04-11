//go:build wireinject
// +build wireinject

// The build constraint above tells the Go compiler to ignore this file;
// Wire reads it instead to generate wire_gen.go.

package app

import (
	"emotionalBeach/config"
	githubhandler "emotionalBeach/internal/handler/github"
	healthhandler "emotionalBeach/internal/handler/health"
	relationhandler "emotionalBeach/internal/handler/relation"
	userhandler "emotionalBeach/internal/handler/user"
	webhookhandler "emotionalBeach/internal/handler/webhook"
	infradb "emotionalBeach/internal/infra/db"
	infralogger "emotionalBeach/internal/infra/logger"
	relationrepo "emotionalBeach/internal/repository/relation"
	userrepo "emotionalBeach/internal/repository/user"
	"emotionalBeach/internal/server"
	githubsvc "emotionalBeach/internal/service/github"
	healthsvc "emotionalBeach/internal/service/health"
	"emotionalBeach/internal/service/notification"
	relationsvc "emotionalBeach/internal/service/relation"
	usersvc "emotionalBeach/internal/service/user"

	"github.com/google/wire"
)

// InitializeApp is the Wire injector.
// Wire reads this function signature, resolves the provider graph, and writes
// the concrete implementation into wire_gen.go.
func InitializeApp(cfg *config.Config) (*App, func(), error) {
	wire.Build(
		// ── Infrastructure ─────────────────────────────────────────────────
		infradb.Set,      // ProvideDB (*gorm.DB, cleanup), ProvideRedis (*redis.Client, cleanup)
		infralogger.Set,  // ProvideLoggers (*Loggers, cleanup)

		// ── Repositories (with interface bindings) ─────────────────────────
		userrepo.Set,     // NewGormRepo + wire.Bind(user.Repository ← *GormRepo)
		relationrepo.Set, // NewGormRepo + wire.Bind(relation.Repository ← *GormRepo)

		// ── Services (with interface bindings) ────────────────────────────
		usersvc.Set,      // NewSvc + wire.Bind(user.Service ← *Svc)
		relationsvc.Set,  // NewSvc + wire.Bind(relation.Service ← *Svc)
		githubsvc.NewSvc,
		healthsvc.NewSvc,
		notification.NewEmailSender,

		// ── HTTP Handlers ──────────────────────────────────────────────────
		userhandler.NewHandler,
		relationhandler.NewHandler,
		githubhandler.NewHandler,
		healthhandler.NewHandler,
		webhookhandler.NewHandler,

		// ── Server ────────────────────────────────────────────────────────
		server.NewRouter,
		server.NewHTTPServer,

		// ── App ───────────────────────────────────────────────────────────
		NewApp,
	)
	return nil, nil, nil
}

