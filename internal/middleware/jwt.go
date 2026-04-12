package middleware

import (
	"emotionalBeach/internal/global"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// uidKey is an unexported context-key type to avoid collisions with other packages.
type uidKey struct{}

var (
	jwtSecret     = []byte("emotionBeach")
	jwtSecretLock sync.RWMutex
)

// ---------------------------------------------------------------------------
// Token parse cache
// ---------------------------------------------------------------------------

type cacheEntry struct {
	claims    *Claims
	expiresAt time.Time
}

var (
	tokenCache     = make(map[string]cacheEntry)
	tokenCacheMu   sync.RWMutex
	cacheOnce      sync.Once
)

// startCacheCleanup launches a background goroutine that evicts stale entries
// every 5 minutes so the cache does not grow without bound.
func startCacheCleanup() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			tokenCacheMu.Lock()
			for tok, e := range tokenCache {
				if now.After(e.expiresAt) {
					delete(tokenCache, tok)
				}
			}
			tokenCacheMu.Unlock()
		}
	}()
}

func cachedClaims(token string) (*Claims, bool) {
	tokenCacheMu.RLock()
	e, ok := tokenCache[token]
	tokenCacheMu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.claims, true
}

func cacheClaims(token string, claims *Claims) {
	if claims.ExpiresAt == nil {
		return
	}
	tokenCacheMu.Lock()
	tokenCache[token] = cacheEntry{claims: claims, expiresAt: claims.ExpiresAt.Time}
	tokenCacheMu.Unlock()
}

func flushCache() {
	tokenCacheMu.Lock()
	tokenCache = make(map[string]cacheEntry)
	tokenCacheMu.Unlock()
}

// ---------------------------------------------------------------------------
// Secret management
// ---------------------------------------------------------------------------

// SetJWTSecret updates the JWT secret from config; empty values are ignored.
// Changing the secret also invalidates the token parse cache.
func SetJWTSecret(secret string) {
	if strings.TrimSpace(secret) == "" {
		return
	}
	jwtSecretLock.Lock()
	jwtSecret = []byte(secret)
	jwtSecretLock.Unlock()
	flushCache()
}

// currentJWTSecret returns a defensive copy of the current secret so callers
// cannot accidentally mutate the shared slice.
func currentJWTSecret() []byte {
	jwtSecretLock.RLock()
	defer jwtSecretLock.RUnlock()
	cp := make([]byte, len(jwtSecret))
	copy(cp, jwtSecret)
	return cp
}

// ---------------------------------------------------------------------------
// Token helpers
// ---------------------------------------------------------------------------

func extractToken(raw string) string {
	raw = strings.TrimSpace(raw)
	if after, ok := strings.CutPrefix(raw, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	if after, ok := strings.CutPrefix(raw, "bearer "); ok {
		return strings.TrimSpace(after)
	}
	return raw
}

// ---------------------------------------------------------------------------
// Claims & token generation
// ---------------------------------------------------------------------------

// Claims holds user state and additional JWT metadata.
type Claims struct {
	UserID uint `json:"userId"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed HS256 JWT for the given user and issuer.
func GenerateToken(userId uint, iss string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			Issuer:    iss,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(currentJWTSecret())
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

func AuthJwt() gin.HandlerFunc {
	// Start the cache-cleanup goroutine exactly once when the middleware is first
	// registered, rather than on every request.
	cacheOnce.Do(startCacheCleanup)

	return func(c *gin.Context) {
		token := extractToken(c.GetHeader("Authorization"))
		if token == "" {
			global.Error(c, http.StatusUnauthorized, "认证信息(Authorization)不能为空!")
			c.Abort()
			return
		}

		claims, err := ParseToken(token)
		if err != nil || claims == nil || claims.UserID == 0 {
			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				global.Error(c, http.StatusUnauthorized, "Token已过期，请重新登录")
			case errors.Is(err, jwt.ErrTokenMalformed):
				global.Error(c, http.StatusUnauthorized, "Token格式无效")
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				global.Error(c, http.StatusUnauthorized, "Token签名无效")
			default:
				global.Error(c, http.StatusUnauthorized, "无效的认证信息")
			}
			zap.S().Infow("token认证失败", zap.Error(err))
			c.Abort()
			return
		}

		// Store with a typed key to avoid string-key collisions.
		c.Set("Uid", claims.UserID) // keep string key for downstream compatibility
		c.Set(uidKey{}, claims.UserID)
		zap.S().Info("token认证成功")
		c.Next()
	}
}

// GetUID retrieves the authenticated user ID from the Gin context.
// It is the type-safe counterpart of c.Get("Uid").
func GetUID(c *gin.Context) (uint, bool) {
	v, exists := c.Get(string("Uid"))
	if !exists {
		return 0, false
	}
	uid, ok := v.(uint)
	return uid, ok
}

// ---------------------------------------------------------------------------
// ParseToken
// ---------------------------------------------------------------------------

// ParseToken validates a JWT string and returns its Claims.
// Results are cached in memory (keyed by raw token) to avoid redundant
// cryptographic operations on repeated requests with the same token.
func ParseToken(tokenStr string) (*Claims, error) {
	// Fast path: return cached result if still valid.
	if claims, ok := cachedClaims(tokenStr); ok {
		return claims, nil
	}

	tokenClaims, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return currentJWTSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	if tokenClaims == nil || !tokenClaims.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := tokenClaims.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	cacheClaims(tokenStr, claims)
	return claims, nil
}
