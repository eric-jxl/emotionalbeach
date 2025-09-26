package middleware

import (
	"emotionalBeach/internal/initialize"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		coloredPath := fmt.Sprintf("\u001B[32m%-*d | %-*s | %-*s\u001B[0m", 3, status, 4, c.Request.Method, 30, path+query)

		initialize.Logger.Info(coloredPath,
			zap.String("client_ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
