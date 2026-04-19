// Package server wires the HTTP server, middleware, and all route handlers.
// Following the arip-samp pattern, a single New() function is the Wire entry
// point. All handler functions in this package share the package-level svc
// variable — no Handler structs, no intermediate router factories.
//
// @title           EmotionalBeach API
// @version         1.0
// @description     emotionalBeach backend REST API
// @termsOfService  http://swagger.io/terms/
// @contact.name    API Support
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html
// @host            localhost:8080
// @BasePath        /v1
package server

import (
	"emotionalBeach/config"
	_ "emotionalBeach/docs"
	"emotionalBeach/internal/common"
	"emotionalBeach/internal/infra"
	"emotionalBeach/internal/middleware"
	"emotionalBeach/internal/service"
	"emotionalBeach/internal/templates"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// svc is set once in New() and shared by all handler functions in this package.
var svc *service.Service

// New constructs a fully-wired *http.Server.
// It is the single provider registered with Wire — no separate router factory,
// no handler constructor calls, matching the arip-samp http.New pattern.
func New(cfg *config.Config, s *service.Service, loggers *infra.Loggers) *http.Server {
	svc = s

	gin.SetMode(gin.ReleaseMode)
	middleware.SetJWTSecret(cfg.Server.JWTSecret)

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		gin.Recovery(),
		middleware.PanicRecoveryMiddleware(),
		middleware.PrometheusMiddleware(),
		middleware.ZapLogger(loggers.Access),
	)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status_code": http.StatusNotFound,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"error":       "请求的资源不存在",
			"request_id":  middleware.GetRequestID(c),
		})
	})

	initRouter(r, cfg)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSec) * time.Second,
	}
}

func initRouter(r *gin.Engine, cfg *config.Config) {
	// ── System endpoints (no auth) ──────────────────────────────────────────
	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	r.GET("/metrics", middleware.MetricsHandler())
	r.GET("/health", healthCheck)

	// ── Static assets ───────────────────────────────────────────────────────
	assetsFS, assetsErr := fs.Sub(templates.AssetHTML, "assets")
	if assetsErr != nil {
		zap.S().Errorf("assets sub-fs failed: %v", assetsErr)
	}
	fileServer := http.FileServer(http.FS(assetsFS))
	r.GET("/assets/*filepath", func(c *gin.Context) {
		if assetsErr != nil {
			common.ServerError(c, "assets unavailable")
			return
		}
		middleware.AssetsCacheMiddleware()(c)
		if c.IsAborted() {
			return
		}
		http.StripPrefix("/assets", fileServer).ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/swagger/*any", swaggerHandler())
	r.GET("/", indexPage)

	// ── Public auth routes ──────────────────────────────────────────────────
	r.GET("/login/github", githubLogin)
	r.GET("/callback", githubCallback)
	r.Any("/login", func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {
			login(c)
		} else {
			common.Fail(c, http.StatusMethodNotAllowed, "Only POST is allowed")
		}
	})
	r.POST("/register", register)

	// ── Token 校验 / 刷新（无需 JWT 中间件，自行校验）──────────────────────
	r.GET("/auth/verify", verifyToken)
	r.POST("/auth/refresh", refreshToken)

	// ── Protected v1 routes (JWT + IP rate-limit) ───────────────────────────
	ipLimiter := middleware.NewIPRateLimiter(rate.Every(10*time.Second), 5)
	v1 := r.Group("/v1", middleware.AuthJwt(), middleware.RateLimitMiddleware(ipLimiter))

	u := v1.Group("user")
	u.GET("/list", getUsers)
	u.GET("/condition", getAppointUser)
	u.POST("/update", updateUser)
	u.DELETE("/delete", deleteUser)

	rel := v1.Group("relation")
	rel.POST("/list", friendList)
	rel.POST("/add", addFriendByName)

	api := v1.Group("api")
	api.POST("/webhook", webhookEmail)
}

func swaggerHandler() gin.HandlerFunc {
	swaggerIndex, err := templates.SwaggerUIHTML.ReadFile("swagger_ui.html")
	if err != nil {
		zap.S().Fatalf("failed to embed swagger_ui.html: %v", err)
	}
	asset := ginSwagger.WrapHandler(swaggerFiles.Handler)
	return func(c *gin.Context) {
		if p := c.Param("any"); p == "/" || p == "/index.html" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerIndex)
			return
		}
		asset(c)
	}
}

func indexPage(c *gin.Context) {
	data, err := templates.IndexHTML.ReadFile("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading index.html")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

