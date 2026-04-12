// Package di wires and manages the application lifecycle.
package di

import (
	"context"
	"emotionalBeach/config"
	"emotionalBeach/internal/infra"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App is the top-level application object.
type App struct {
	cfg     *config.Config
	server  *http.Server
	loggers *infra.Loggers
}

// NewApp assembles the application. Wire calls this last in the provider graph.
// Startup side-effects (migration, cache warm-up, Prometheus collectors) are
// handled via infra helpers so this function stays declarative.
func NewApp(
	cfg *config.Config,
	srv *http.Server,
	db *gorm.DB,
	rdb *redis.Client,
	loggers *infra.Loggers,
) *App {
	infra.AutoMigrate(db)          // exits process if -migrate flag is set
	infra.CachePreload(cfg, rdb, db) // no-op when Redis disabled
	infra.RegisterCollectors(db)   // Prometheus DB-pool scrape collector
	return &App{cfg: cfg, server: srv, loggers: loggers}
}

// Run starts the HTTP server, blocks until a termination signal, then shuts down.
// DB and Redis are closed by Wire's cleanup func (defer cleanup() in main.go).
func (a *App) Run() {
	// Start runtime metrics updater; stop it on exit.
	stopMetrics := infra.StartRuntimeCollector(15 * time.Second)
	defer close(stopMetrics)

	go func() {
		zap.S().Infof("🚀 HTTP server listening on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Errorf("HTTP server error: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	zap.S().Infof("⏳ received signal %s, starting graceful shutdown…", <-quit)

	timeout := time.Duration(a.cfg.Server.ShutdownTimeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		zap.S().Errorf("HTTP shutdown error: %v", err)
	}
	if a.loggers != nil {
		_ = a.loggers.Sys.Sync()
		_ = a.loggers.Access.Sync()
	}
	zap.S().Info("✅ server exited cleanly")
}
