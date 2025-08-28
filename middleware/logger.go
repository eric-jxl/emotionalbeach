package middleware

import (
	"emotionalBeach/initialize"
	"go.uber.org/zap"
	"time"

	"github.com/gin-gonic/gin"
)

// ZapLogger 自定义 zap logger 中间件
func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		coloredPath := "\033[32m" + path + "\033[0m"

		initialize.Logger.Info(coloredPath,
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("query", query),
			zap.String("client_ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
