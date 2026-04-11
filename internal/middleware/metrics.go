package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ──────────────────────────────────────────────────────────────
// Prometheus 指标定义
// Namespace "eb" = EmotionalBeach
// ──────────────────────────────────────────────────────────────

var (
	// HTTP 请求总量（method × path × status_code）
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "eb",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests processed.",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTP 请求耗时直方图（method × path）
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "eb",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency in seconds.",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// 当前正在处理的请求数
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "eb",
			Name:      "http_requests_in_flight",
			Help:      "Current number of HTTP requests being handled.",
		},
	)
)

// PrometheusMiddleware 记录每个 HTTP 请求的 Prometheus 指标。
// 使用 c.FullPath() 获取路由模板（如 /v1/user/:id），避免标签基数爆炸。
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// /metrics 端点自身不计入指标，避免自我干扰
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		httpRequestsInFlight.Inc()
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		httpRequestsInFlight.Dec()

		path := c.FullPath()
		if path == "" {
			path = "unmatched" // 404 路由
		}

		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// MetricsHandler 返回 Prometheus 抓取端点的 Gin HandlerFunc。
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RecordPanic 在 Recovery 拦截 panic 后调用，递增专属计数器。
var panicRecoveryTotal = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "eb",
	Name:      "http_panics_recovered_total",
	Help:      "Total number of HTTP handler panics recovered.",
})

// PanicRecoveryMiddleware 在 gin.Recovery 基础上补充 Prometheus 计数 + 结构化日志。
func PanicRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				panicRecoveryTotal.Inc()
				// 仍通过标准 gin Recovery 处理，但同时 500
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

