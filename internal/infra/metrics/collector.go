package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

// ── DB Pool Collector ─────────────────────────────────────────────────────────
// DBPoolCollector implements prometheus.Collector and reads sql.DBStats on
// every scrape, so the values are always fresh without background polling.

type DBPoolCollector struct {
	db          *gorm.DB
	openConns   *prometheus.Desc
	inUseConns  *prometheus.Desc
	idleConns   *prometheus.Desc
	waitCount   *prometheus.Desc
	waitSeconds *prometheus.Desc
	maxOpen     *prometheus.Desc
	maxIdle     *prometheus.Desc
}

// NewDBPoolCollector constructs and registers a DBPoolCollector.
// Call this once at application start (e.g. inside NewApp).
func NewDBPoolCollector(db *gorm.DB) *DBPoolCollector {
	c := &DBPoolCollector{
		db: db,
		openConns: prometheus.NewDesc(
			ns+"_db_open_connections",
			"Current number of open connections to the database (in-use + idle).",
			nil, nil,
		),
		inUseConns: prometheus.NewDesc(
			ns+"_db_in_use_connections",
			"Number of connections currently in use by the application.",
			nil, nil,
		),
		idleConns: prometheus.NewDesc(
			ns+"_db_idle_connections",
			"Number of idle connections in the pool.",
			nil, nil,
		),
		waitCount: prometheus.NewDesc(
			ns+"_db_wait_count_total",
			"Total number of times a goroutine had to wait for a free connection.",
			nil, nil,
		),
		waitSeconds: prometheus.NewDesc(
			ns+"_db_wait_duration_seconds_total",
			"Total time (seconds) spent waiting for a free connection.",
			nil, nil,
		),
		maxOpen: prometheus.NewDesc(
			ns+"_db_max_open_connections",
			"Maximum number of open connections configured.",
			nil, nil,
		),
		maxIdle: prometheus.NewDesc(
			ns+"_db_max_idle_connections",
			"Maximum number of idle connections configured.",
			nil, nil,
		),
	}
	prometheus.MustRegister(c)
	return c
}

func (c *DBPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.openConns
	ch <- c.inUseConns
	ch <- c.idleConns
	ch <- c.waitCount
	ch <- c.waitSeconds
	ch <- c.maxOpen
	ch <- c.maxIdle
}

func (c *DBPoolCollector) Collect(ch chan<- prometheus.Metric) {
	sqlDB, err := c.db.DB()
	if err != nil {
		return
	}
	s := sqlDB.Stats()
	emit := func(desc *prometheus.Desc, t prometheus.ValueType, v float64) {
		ch <- prometheus.MustNewConstMetric(desc, t, v)
	}
	emit(c.openConns, prometheus.GaugeValue, float64(s.OpenConnections))
	emit(c.inUseConns, prometheus.GaugeValue, float64(s.InUse))
	emit(c.idleConns, prometheus.GaugeValue, float64(s.Idle))
	emit(c.waitCount, prometheus.CounterValue, float64(s.WaitCount))
	emit(c.waitSeconds, prometheus.CounterValue, s.WaitDuration.Seconds())
	emit(c.maxOpen, prometheus.GaugeValue, float64(s.MaxOpenConnections))
	emit(c.maxIdle, prometheus.GaugeValue, float64(s.MaxIdleClosed))
}

// ── Runtime metrics updater ───────────────────────────────────────────────────
// StartRuntimeCollector launches a goroutine that refreshes runtime gauges
// every interval. Call once from NewApp.

func StartRuntimeCollector(interval time.Duration, stop <-chan struct{}) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				MemAllocMB.Set(float64(m.Alloc) / 1024 / 1024)
			case <-stop:
				return
			}
		}
	}()
}
