package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查请求来源是否在允许列表中
		allowThisOrigin := false
		for _, o := range allowedOrigins {
			if strings.EqualFold(o, origin) {
				allowThisOrigin = true
				break
			}
		}

		if allowThisOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		// 允许跨域携带 Cookie / sessionId
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
