package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey 是在 Gin context 和响应头中存放 Request ID 的键名。
const RequestIDKey = "X-Request-Id"

// RequestID 中间件：每个请求生成唯一 ID，优先复用客户端传入值，
// 并写入响应头和 Gin context，供日志、下游追踪使用。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(RequestIDKey)
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set(RequestIDKey, rid)
		c.Header(RequestIDKey, rid)
		c.Next()
	}
}

// GetRequestID 从 Gin context 中安全提取 Request ID。
func GetRequestID(c *gin.Context) string {
	if rid, exists := c.Get(RequestIDKey); exists {
		if s, ok := rid.(string); ok {
			return s
		}
	}
	return ""
}

