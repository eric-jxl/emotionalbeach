package infra

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
)

const ns = "eb"

// ── User ─────────────────────────────────────────────────────────────────────

var UserRegistrationsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: ns, Subsystem: "user", Name: "registrations_total",
	Help: "Total number of successful user registrations.",
})

var UserLoginsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: ns, Subsystem: "user", Name: "logins_total",
	Help: "Total login attempts labelled by outcome.",
}, []string{"status"})

// ── Relation ─────────────────────────────────────────────────────────────────

var FriendAddTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: ns, Subsystem: "relation", Name: "friend_add_total",
	Help: "Total friend-add operations labelled by outcome.",
}, []string{"status"})

// ── Notification ─────────────────────────────────────────────────────────────

var EmailSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: ns, Subsystem: "notification", Name: "emails_sent_total",
	Help: "Total email send attempts labelled by outcome.",
}, []string{"status"})

var EmailReceiversTotal = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: ns, Subsystem: "notification", Name: "email_receivers_total",
	Help: "Total number of email recipients across all successful sends.",
})

// ── Health gauges ─────────────────────────────────────────────────────────────

var ComponentHealthGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: ns, Name: "component_health",
	Help: "Health status of each component: 1 = ok, 0 = error/degraded.",
}, []string{"component"})

var ComponentLatencyMs = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: ns, Name: "component_probe_latency_ms",
	Help: "Last health-probe round-trip latency in milliseconds.",
}, []string{"component"})

// ── Runtime ───────────────────────────────────────────────────────────────────

var MemAllocMB = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: ns, Name: "runtime_mem_alloc_mb",
	Help: "Current heap memory allocated by the Go runtime in MiB.",
})

// StartRuntimeCollector launches a goroutine that refreshes runtime gauges
// every interval. Returns a stop channel; close it to halt the collector.
func StartRuntimeCollector(interval time.Duration) chan struct{} {
	stop := make(chan struct{})
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
	return stop
}

// ── DB Pool Collector ─────────────────────────────────────────────────────────

type dbPoolCollector struct {
	db                            *gorm.DB
	openConns, inUseConns, idleConns,
	waitCount, waitSeconds, maxOpen, maxIdle *prometheus.Desc
}

func newDBPoolCollector(db *gorm.DB) {
	c := &dbPoolCollector{
		db:          db,
		openConns:   prometheus.NewDesc(ns+"_db_open_connections", "Current number of open connections.", nil, nil),
		inUseConns:  prometheus.NewDesc(ns+"_db_in_use_connections", "Connections currently in use.", nil, nil),
		idleConns:   prometheus.NewDesc(ns+"_db_idle_connections", "Idle connections in the pool.", nil, nil),
		waitCount:   prometheus.NewDesc(ns+"_db_wait_count_total", "Times a goroutine waited for a free connection.", nil, nil),
		waitSeconds: prometheus.NewDesc(ns+"_db_wait_duration_seconds_total", "Total time waiting for a connection (s).", nil, nil),
		maxOpen:     prometheus.NewDesc(ns+"_db_max_open_connections", "Maximum open connections configured.", nil, nil),
		maxIdle:     prometheus.NewDesc(ns+"_db_max_idle_connections", "Maximum idle connections configured.", nil, nil),
	}
	prometheus.MustRegister(c)
}

func (c *dbPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range []*prometheus.Desc{
		c.openConns, c.inUseConns, c.idleConns,
		c.waitCount, c.waitSeconds, c.maxOpen, c.maxIdle,
	} {
		ch <- d
	}
}

func (c *dbPoolCollector) Collect(ch chan<- prometheus.Metric) {
	sqlDB, err := c.db.DB()
	if err != nil {
		return
	}
	s := sqlDB.Stats()
	emit := func(d *prometheus.Desc, t prometheus.ValueType, v float64) {
		ch <- prometheus.MustNewConstMetric(d, t, v)
	}
	emit(c.openConns, prometheus.GaugeValue, float64(s.OpenConnections))
	emit(c.inUseConns, prometheus.GaugeValue, float64(s.InUse))
	emit(c.idleConns, prometheus.GaugeValue, float64(s.Idle))
	emit(c.waitCount, prometheus.CounterValue, float64(s.WaitCount))
	emit(c.waitSeconds, prometheus.CounterValue, s.WaitDuration.Seconds())
	emit(c.maxOpen, prometheus.GaugeValue, float64(s.MaxOpenConnections))
	emit(c.maxIdle, prometheus.GaugeValue, float64(s.MaxIdleClosed))
}

