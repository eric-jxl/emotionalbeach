package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// slowRequestThreshold 超过此阈值的请求以 Warn 级别记录，并标注 [SLOW]。
const slowRequestThreshold = 500 * time.Millisecond

// ── ANSI 颜色码 ──────────────────────────────────────────────────────────────
const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
	ansiCyan   = "\033[36m"
	ansiBold   = "\033[1m"
)

func statusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return ansiGreen
	case code >= 300 && code < 400:
		return ansiCyan
	case code >= 400 && code < 500:
		return ansiYellow
	default:
		return ansiRed
	}
}

func methodColor(method string) string {
	switch method {
	case "GET":
		return ansiBlue
	case "POST":
		return ansiGreen
	case "PUT", "PATCH":
		return ansiCyan
	case "DELETE":
		return ansiRed
	default:
		return ansiYellow
	}
}

// fmtLatency 把 duration 格式化为更易读的短字符串。
func fmtLatency(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.3fµs", float64(d.Nanoseconds())/1e3)
	case d < time.Second:
		return fmt.Sprintf("%.3fms", float64(d.Nanoseconds())/1e6)
	default:
		return fmt.Sprintf("%.3fs", d.Seconds())
	}
}

// ZapLogger returns an HTTP access-log middleware backed by the provided logger.
//
// Console output (single coloured line, no key=value noise):
//
//	2025-01-01 15:04:05  INFO   200 | GET     /v1/user/list  |   5.234ms | 127.0.0.1 | <request-id>
//
// JSON file output (structured fields for ELK/Loki ingestion).
func ZapLogger(accessLogger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		rid := GetRequestID(c)

		// 拼接显示路径（含查询字符串）
		displayPath := path
		if query != "" {
			displayPath = path + "?" + query
		}

		isSlow := latency > slowRequestThreshold && status < 400

		// 慢请求强制使用黄色高亮状态码
		sc := statusColor(status)
		if isSlow {
			sc = ansiYellow
		}

		// ── 控制台格式化行（由 messageOnlyCore 保证此为唯一输出，无附加字段）
		//    格式参考 Gin 默认输出，增加 request_id 列
		//    示例：200 | GET     /v1/user/list                      |   5.234ms | 127.0.0.1       | abc-uuid
		slowMark := ""
		if isSlow {
			slowMark = ansiBold + ansiYellow + " [SLOW]" + ansiReset
		}
		consoleLine := fmt.Sprintf(
			"%s%3d%s | %s%-7s%s %-45s | %10s | %-16s | %s%s",
			sc, status, ansiReset,
			methodColor(method), method, ansiReset,
			displayPath,
			fmtLatency(latency),
			clientIP,
			rid,
			slowMark,
		)

		// ── JSON 文件字段（不在控制台显示，由 messageOnlyCore 过滤）──────
		fields := []zap.Field{
			zap.String("request_id", rid),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.String("client_ip", clientIP),
			zap.Int64("latency_us", latency.Microseconds()),
			zap.Int("body_size", c.Writer.Size()),
			zap.String("user_agent", c.Request.UserAgent()),
		}
		if errs := c.Errors.ByType(gin.ErrorTypePrivate).String(); errs != "" {
			fields = append(fields, zap.String("errors", errs))
		}

		// ── Log at the right level based on status / latency ─────────────
		switch {
		case isSlow:
			accessLogger.Warn(consoleLine, fields...)
		case status >= 500:
			accessLogger.Error(consoleLine, fields...)
		case status >= 400:
			accessLogger.Warn(consoleLine, fields...)
		default:
			accessLogger.Info(consoleLine, fields...)
		}
	}
}
