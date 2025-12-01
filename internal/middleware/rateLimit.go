package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter RateLimiter 中间件使用的限流器管理
type IPRateLimiter struct {
	limiters sync.Map
	r        rate.Limit
	b        int
}

// NewIPRateLimiter 创建一个新的 IP 限流器管理器
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		r: r,
		b: b,
	}
}

// AddIP 为给定的 IP 添加一个限流器（如果不存在）
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	limiter := rate.NewLimiter(i.r, i.b)
	i.limiters.Store(ip, limiter)
	return limiter
}

// GetLimiter 获取与给定 IP 关联的限流器，如果不存在则创建一个
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	if limiter, ok := i.limiters.Load(ip); ok {
		return limiter.(*rate.Limiter)
	}
	return i.AddIP(ip)
}
func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		limiterForIP := limiter.GetLimiter(clientIP)

		// 尝试获取一个令牌，不阻塞
		if !limiterForIP.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": http.StatusTooManyRequests, "error": "请求过于频繁，请稍后再试"})
			return
		}

		c.Next()
	}
}
