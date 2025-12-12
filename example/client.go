package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func RequestClient() *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,  // TCP连接建立超时
			KeepAlive: 30 * time.Second, // 连接保活时间
		}).DialContext,
		ResponseHeaderTimeout: 5 * time.Second, // 等待响应头超时
		MaxIdleConnsPerHost:   100,             // 每个主机的最大空闲连接
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second, // 整个请求的超时时间
	}
	return client
}

//4.2 带追踪的超时控制

func requestWithTracing(ctx context.Context) (*http.Response, error) {
	// 从父上下文派生5秒超时的子上下文
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // 确保无论成功失败都取消上下文

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加分布式追踪信息
	req.Header.Set("X-Request-ID", ctx.Value("request-id").(string))

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 2 * time.Second,
			}).DialContext,
		},
		// 注意: 此处不设置Timeout，完全由context控制
	}

	resp, err := client.Do(req)
	if err != nil {
		// 区分上下文取消和其他错误
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("请求超时: %w", ctx.Err())
		}
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	return resp, nil
}

// RetryPolicy 5.1 指数退避与抖动
type RetryPolicy struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	JitterFactor   float64 // 抖动系数，建议0.1-0.5
}

// Backoff 带抖动的指数退避
func (rp *RetryPolicy) Backoff(attempt int) time.Duration {
	if attempt <= 0 {
		return rp.InitialBackoff
	}
	// 指数增长: InitialBackoff * 2^(attempt-1)
	backoff := rp.InitialBackoff * (1 << (attempt - 1))
	if backoff > rp.MaxBackoff {
		backoff = rp.MaxBackoff
	}
	// 添加抖动: [backoff*(1-jitter), backoff*(1+jitter)]
	jitter := time.Duration(rand.Float64() * float64(backoff) * rp.JitterFactor)
	return backoff - jitter + 2*jitter // 均匀分布在抖动范围内
}

// Retry 通用重试执行器
func Retry(ctx context.Context, policy RetryPolicy, fn func() error) error {
	var err error
	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		if attempt > 0 {
			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				return fmt.Errorf("重试被取消: %w", ctx.Err())
			default:
			}

			backoff := policy.Backoff(attempt)
			timer := time.NewTimer(backoff)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return fmt.Errorf("重试被取消: %w", ctx.Err())
			}
		}

		err = fn()
		if err == nil {
			return nil
		}

		// 判断是否应该重试
	}
	return fmt.Errorf("达到最大重试次数 %d: %w", policy.MaxRetries, err)
}

// IdempotentClient 6.1 请求ID+Redis实现
type IdempotentClient struct {
	redisClient *redis.Client
	prefix      string        // Redis键前缀
	ttl         time.Duration // 幂等键过期时间
}

// NewRequestID 生成唯一请求ID
func (ic *IdempotentClient) NewRequestID() string {
	return uuid.New().String()
}

// Do 执行幂等请求
func (ic *IdempotentClient) Do(req *http.Request, requestID string) (*http.Response, error) {
	// 检查请求是否已处理
	key := fmt.Sprintf("%s:%s", ic.prefix, requestID)
	exists, err := ic.redisClient.Exists(req.Context(), key).Result()
	if err != nil {
		return nil, fmt.Errorf("幂等检查失败: %v", err)
	}
	if exists == 1 {
		// 返回缓存的响应或标记为重复请求
		return nil, fmt.Errorf("请求已处理: %s", requestID)
	}

	// 使用SET NX确保只有一个请求能通过检查
	set, err := ic.redisClient.SetNX(
		req.Context(),
		key,
		"processing",
		ic.ttl,
	).Result()
	if err != nil {
		return nil, fmt.Errorf("幂等锁失败: %v", err)
	}
	if !set {
		return nil, fmt.Errorf("并发请求冲突: %s", requestID)
	}

	// 执行请求
	client := &http.Client{ /* 配置 */ }
	resp, err := client.Do(req)
	if err != nil {
		// 请求失败时删除幂等标记
		ic.redisClient.Del(req.Context(), key)
		return nil, err
	}

	// 请求成功，更新幂等标记状态
	ic.redisClient.Set(req.Context(), key, "completed", ic.ttl)
	return resp, nil
}

// NewOptimizedTransport 7.1 连接池配置
func NewOptimizedTransport() *http.Transport {
	return &http.Transport{
		// 连接池配置
		MaxIdleConns:        1000,             // 全局最大空闲连接
		MaxIdleConnsPerHost: 100,              // 每个主机的最大空闲连接
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间

		// TCP配置
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		// TLS配置
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},

		// 其他优化
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    false, // 启用压缩
	}
}

var requestPool = sync.Pool{
	New: func() interface{} {
		return &http.Request{
			Header: make(http.Header),
		}
	},
}

// AcquireRequest 7.2 sync.Pool内存复用
// 从池获取请求对象
func AcquireRequest() *http.Request {
	req := requestPool.Get().(*http.Request)
	// 重置必要字段
	req.Method = ""
	req.URL = nil
	req.Body = nil
	req.ContentLength = 0
	req.Header = make(http.Header)
	return req
}

// ReleaseRequest 释放请求对象到池
func ReleaseRequest(req *http.Request) {
	requestPool.Put(req)
}
