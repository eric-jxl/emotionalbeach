// Package metrics centralises every application-level Prometheus metric.
//
// Naming convention (aligned with Prometheus best practices):
//   eb_<subsystem>_<name>_<unit>   (namespace = "eb" for EmotionalBeach)
//
// All metrics use promauto so they self-register on import — no explicit
// prometheus.Register() call is needed anywhere.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const ns = "eb"

// ── User ─────────────────────────────────────────────────────────────────────

// UserRegistrationsTotal counts successful user registrations.
var UserRegistrationsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: ns,
	Subsystem: "user",
	Name:      "registrations_total",
	Help:      "Total number of successful user registrations.",
})

// UserLoginsTotal counts login attempts split by outcome.
// Labels: status = "success" | "not_found" | "wrong_password" | "token_error"
var UserLoginsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: ns,
	Subsystem: "user",
	Name:      "logins_total",
	Help:      "Total login attempts labelled by outcome.",
}, []string{"status"})

// ── Relation ─────────────────────────────────────────────────────────────────

// FriendAddTotal counts friend-add operations split by outcome.
// Labels: status = "success" | "already_exists" | "self" | "not_found" | "error"
var FriendAddTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: ns,
	Subsystem: "relation",
	Name:      "friend_add_total",
	Help:      "Total friend-add operations labelled by outcome.",
}, []string{"status"})

// ── Notification ─────────────────────────────────────────────────────────────

// EmailSentTotal counts email send attempts split by outcome.
// Labels: status = "success" | "failure"
var EmailSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: ns,
	Subsystem: "notification",
	Name:      "emails_sent_total",
	Help:      "Total email send attempts labelled by outcome.",
}, []string{"status"})

// EmailReceiversTotal counts total number of recipients across all sends.
var EmailReceiversTotal = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: ns,
	Subsystem: "notification",
	Name:      "email_receivers_total",
	Help:      "Total number of email recipients across all successful sends.",
})

// ── Component health gauges ───────────────────────────────────────────────────
// Value: 1 = healthy, 0 = degraded/error.
// These are written by the health service on every /health probe so that
// Grafana / Alertmanager can alert without polling the HTTP endpoint directly.

// ComponentHealthGauge exposes a per-component health status.
// Labels: component = "database" | "redis"
var ComponentHealthGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: ns,
	Name:      "component_health",
	Help:      "Health status of each component: 1 = ok, 0 = error/degraded.",
}, []string{"component"})

// ComponentLatencyMs records the last probe latency in milliseconds.
// Labels: component = "database" | "redis"
var ComponentLatencyMs = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: ns,
	Name:      "component_probe_latency_ms",
	Help:      "Last health-probe round-trip latency in milliseconds.",
}, []string{"component"})

// ── Runtime ───────────────────────────────────────────────────────────────────
// Note: go_goroutines, go_gc_*, process_* are already exported by the default
// Go runtime and process collectors registered automatically by client_golang.
// We expose a memory gauge here for convenience in dashboards.

// MemAllocMB tracks current heap memory allocation in MiB.
var MemAllocMB = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: ns,
	Name:      "runtime_mem_alloc_mb",
	Help:      "Current heap memory allocated by the Go runtime in MiB.",
})

