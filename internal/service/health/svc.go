// Package healthsvc implements deep health-check logic for DB and Redis.
package healthsvc

import (
	"context"
	"database/sql"
	ebmetrics "emotionalBeach/internal/infra/metrics"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ComponentStatus captures the liveness state of one subsystem.
type ComponentStatus struct {
	Status    string  `json:"status"`
	Message   string  `json:"message,omitempty"`
	LatencyMs float64 `json:"latency_ms,omitempty"`
}

// DBPoolStats mirrors sql.DBStats for JSON serialisation.
type DBPoolStats struct {
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	MaxOpen            int           `json:"max_open"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       string        `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
}

// MemStats holds key runtime memory figures.
type MemStats struct {
	AllocMB   float64 `json:"alloc_mb"`
	SysMB     float64 `json:"sys_mb"`
	NumGC     uint32  `json:"num_gc"`
}

// Report is the top-level health-check response payload.
type Report struct {
	Status     string                     `json:"status"`
	Uptime     string                     `json:"uptime"`
	GoVersion  string                     `json:"go_version"`
	Goroutines int                        `json:"goroutines"`
	Memory     MemStats                   `json:"memory"`
	DBPool     *DBPoolStats               `json:"db_pool,omitempty"`
	Components map[string]ComponentStatus `json:"components"`
}

// Svc runs deep health checks against every registered dependency.
type Svc struct {
	db        *gorm.DB
	rdb       *redis.Client
	startTime time.Time
}

// NewSvc constructs a Svc. Both db and rdb may be nil when disabled.
func NewSvc(db *gorm.DB, rdb *redis.Client) *Svc {
	return &Svc{db: db, rdb: rdb, startTime: time.Now()}
}

// Check runs all component probes and returns a consolidated Report.
// It also updates Prometheus gauges so Grafana can alert without polling /health.
func (s *Svc) Check() Report {
	components := make(map[string]ComponentStatus)
	overallOK := true

	// ── Database probe ───────────────────────────────────────────────────────
	dbStatus := s.probeDB()
	components["database"] = dbStatus
	if dbStatus.Status != "ok" && dbStatus.Status != "unconfigured" {
		overallOK = false
	}

	// ── Redis probe ──────────────────────────────────────────────────────────
	redisStatus := s.probeRedis()
	components["redis"] = redisStatus
	if redisStatus.Status != "ok" && redisStatus.Status != "disabled" {
		overallOK = false
	}

	// ── Emit Prometheus health gauges ────────────────────────────────────────
	setHealthGauge("database", dbStatus)
	setHealthGauge("redis", redisStatus)

	// ── Runtime memory ────────────────────────────────────────────────────────
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	mem := MemStats{
		AllocMB: float64(ms.Alloc) / 1024 / 1024,
		SysMB:   float64(ms.Sys) / 1024 / 1024,
		NumGC:   ms.NumGC,
	}
	ebmetrics.MemAllocMB.Set(mem.AllocMB)

	status := "ok"
	if !overallOK {
		status = "degraded"
		zap.S().Warnw("health check degraded", "components", components)
	}

	return Report{
		Status:     status,
		Uptime:     time.Since(s.startTime).Round(time.Second).String(),
		GoVersion:  runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		Memory:     mem,
		DBPool:     s.dbPoolStats(),
		Components: components,
	}
}

// ── private helpers ───────────────────────────────────────────────────────────

func (s *Svc) probeDB() ComponentStatus {
	if s.db == nil {
		return ComponentStatus{Status: "unconfigured"}
	}
	start := time.Now()
	sqlDB, err := s.db.DB()
	if err != nil {
		return ComponentStatus{Status: "error", Message: err.Error()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err = sqlDB.PingContext(ctx); err != nil {
		return ComponentStatus{Status: "error", Message: err.Error()}
	}
	return ComponentStatus{
		Status:    "ok",
		LatencyMs: ms(start),
	}
}

func (s *Svc) probeRedis() ComponentStatus {
	if s.rdb == nil {
		return ComponentStatus{Status: "disabled"}
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return ComponentStatus{Status: "error", Message: err.Error()}
	}
	return ComponentStatus{
		Status:    "ok",
		LatencyMs: ms(start),
	}
}

func (s *Svc) dbPoolStats() *DBPoolStats {
	if s.db == nil {
		return nil
	}
	sqlDB, err := s.db.DB()
	if err != nil {
		return nil
	}
	st := sqlDB.Stats()
	return poolStatsFrom(st)
}

func poolStatsFrom(s sql.DBStats) *DBPoolStats {
	return &DBPoolStats{
		OpenConnections:   s.OpenConnections,
		InUse:             s.InUse,
		Idle:              s.Idle,
		MaxOpen:           s.MaxOpenConnections,
		WaitCount:         s.WaitCount,
		WaitDuration:      s.WaitDuration.String(),
		MaxIdleClosed:     s.MaxIdleClosed,
		MaxLifetimeClosed: s.MaxLifetimeClosed,
	}
}

// setHealthGauge updates the Prometheus component health gauge and latency gauge.
func setHealthGauge(component string, cs ComponentStatus) {
	v := 0.0
	if cs.Status == "ok" {
		v = 1.0
	}
	ebmetrics.ComponentHealthGauge.WithLabelValues(component).Set(v)
	if cs.LatencyMs > 0 {
		ebmetrics.ComponentLatencyMs.WithLabelValues(component).Set(cs.LatencyMs)
	}
}

func ms(start time.Time) float64 {
	return float64(time.Since(start).Microseconds()) / 1000.0
}
