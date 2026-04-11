// Package app contains the application lifecycle manager.
package app

import (
	"context"
	"emotionalBeach/config"
	"emotionalBeach/internal/infra/cache"
	infralogger "emotionalBeach/internal/infra/logger"
	ebmetrics "emotionalBeach/internal/infra/metrics"
	"emotionalBeach/internal/models"
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
// It owns every infrastructure resource and controls the HTTP server lifecycle.
type App struct {
	cfg          *config.Config
	server       *http.Server
	db           *gorm.DB
	redis        *redis.Client
	loggers      *infralogger.Loggers
	metricsStop  chan struct{} // signals background metric collectors to stop
}

// NewApp assembles the application. Wire calls this last in the provider graph.
// Migration, cache warm-up, and Prometheus collector startup are handled here.
func NewApp(
	cfg *config.Config,
	srv *http.Server,
	db *gorm.DB,
	rdb *redis.Client,
	loggers *infralogger.Loggers,
) *App {
	// ── Database migration (opt-in via -migrate / --migrate flag) ─────────
	if hasMigrateArg(os.Args) {
		if err := db.AutoMigrate(&models.UserBasic{}, &models.Relation{}); err != nil {
			zap.S().Fatalf("❌ migration failed: %v", err)
		}
		zap.S().Info("✅ migration done")
		os.Exit(0)
	}

	// ── Redis cache warm-up ────────────────────────────────────────────────
	if cfg.Server.EnableRedis && rdb != nil {
		if err := cache.Preload(rdb, db); err != nil {
			zap.S().Fatalf("❌ Redis preload failed: %v", err)
		}
		zap.S().Info("✅ Redis cache warmed up")
	}

	// ── Prometheus: DB pool collector (scrape-driven, no polling) ─────────
	ebmetrics.NewDBPoolCollector(db)

	// ── Prometheus: runtime memory gauge (updated every 15 s) ─────────────
	stop := make(chan struct{})
	ebmetrics.StartRuntimeCollector(15*time.Second, stop)

	return &App{cfg: cfg, server: srv, db: db, redis: rdb, loggers: loggers, metricsStop: stop}
}

// Run starts the HTTP server and blocks until a termination signal is received,
// then performs a graceful shutdown of all resources.
func (a *App) Run() {
	// Start HTTP server in background.
	go func() {
		zap.S().Infof("🚀 HTTP server listening on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Errorf("HTTP server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for OS interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	zap.S().Infof("⏳ received signal %s, starting graceful shutdown…", sig)

	a.shutdown()
	zap.S().Info("✅ server exited cleanly")
}

// shutdown drains the HTTP server then closes every infrastructure connection.
func (a *App) shutdown() {
	timeout := time.Duration(a.cfg.Server.ShutdownTimeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 1. Stop accepting new requests; drain in-flight ones.
	if err := a.server.Shutdown(ctx); err != nil {
		zap.S().Errorf("HTTP shutdown error: %v", err)
	}

	// 2. Stop background metric collectors.
	if a.metricsStop != nil {
		close(a.metricsStop)
	}

	// 3. Close database connection pool.
	if a.db != nil {
		if sqlDB, err := a.db.DB(); err == nil {
			if err = sqlDB.Close(); err != nil {
				zap.S().Warnf("DB close error: %v", err)
			}
		}
	}

	// 4. Close Redis client.
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			zap.S().Warnf("Redis close error: %v", err)
		}
	}

	// 5. Flush log buffers.
	if a.loggers != nil {
		_ = a.loggers.Sys.Sync()
		_ = a.loggers.Access.Sync()
	}
}

func hasMigrateArg(args []string) bool {
	for _, a := range args {
		if a == "-migrate" || a == "--migrate" {
			return true
		}
	}
	return false
}

