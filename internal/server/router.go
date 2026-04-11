// Package server composes all domain handlers into a single *gin.Engine.
// @title Swagger Example API
// @version 1.0
// @description This is a sample server for a hypothetical API.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /v1
package server

import (
	"emotionalBeach/config"
	_ "emotionalBeach/docs"
	githubhandler "emotionalBeach/internal/handler/github"
	healthhandler "emotionalBeach/internal/handler/health"
	relationhandler "emotionalBeach/internal/handler/relation"
	userhandler "emotionalBeach/internal/handler/user"
	webhookhandler "emotionalBeach/internal/handler/webhook"
	infralogger "emotionalBeach/internal/infra/logger"
	"emotionalBeach/internal/middleware"
	"emotionalBeach/internal/templates"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// NewRouter composes all routes and returns the *gin.Engine.
// Every dependency is injected — no global state is accessed here.
func NewRouter(
	cfg *config.Config,
	loggers *infralogger.Loggers,
	userH *userhandler.Handler,
	relationH *relationhandler.Handler,
	githubH *githubhandler.Handler,
	healthH *healthhandler.Handler,
	webhookH *webhookhandler.Handler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Apply JWT secret from config (called once at startup — safe).
	middleware.SetJWTSecret(cfg.Server.JWTSecret)

	// ── Global middleware (order matters) ──────────────────────────────────
	router.Use(
		middleware.RequestID(),               // 1. inject X-Request-Id
		gin.Recovery(),                       // 2. panic recovery
		middleware.PanicRecoveryMiddleware(), // 3. panic counter + structured response
		middleware.PrometheusMiddleware(),    // 4. Prometheus metrics
		middleware.ZapLogger(loggers.Access), // 5. structured access log (injected logger)
	)

	// ── 404 handler ────────────────────────────────────────────────────────
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status_code": http.StatusNotFound,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"error":       "请求的资源不存在",
			"request_id":  middleware.GetRequestID(c),
		})
	})

	// ── System endpoints (no rate-limit) ───────────────────────────────────
	router.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	router.GET("/metrics", middleware.MetricsHandler())
	healthH.RegisterRoutes(router)

	// ── Static assets ──────────────────────────────────────────────────────
	assetsFS, assetsErr := fs.Sub(templates.AssetHTML, "assets")
	if assetsErr != nil {
		zap.S().Errorf("assets sub-fs failed: %v", assetsErr)
	}
	fileServer := http.FileServer(http.FS(assetsFS))
	router.GET("/assets/*filepath", func(c *gin.Context) {
		if assetsErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "assets unavailable"})
			return
		}
		middleware.AssetsCacheMiddleware()(c)
		if c.IsAborted() {
			return
		}
		http.StripPrefix("/assets", fileServer).ServeHTTP(c.Writer, c.Request)
	})

	router.GET("/swagger/*any", buildSwaggerHandler())
	router.GET("/", func(c *gin.Context) {
		data, err := templates.IndexHTML.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error loading index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// ── Public auth routes ─────────────────────────────────────────────────
	githubH.RegisterRoutes(router)
	router.Any("/login", func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {
			userH.Login(c)
		} else {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"code": http.StatusMethodNotAllowed, "message": "Only POST is allowed"})
		}
	})
	router.POST("/register", userH.Register)

	// ── Protected v1 routes (JWT + IP rate-limit) ──────────────────────────
	ipLimiter := middleware.NewIPRateLimiter(rate.Every(10*time.Second), 5)
	v1 := router.Group("/v1", middleware.AuthJwt(), middleware.RateLimitMiddleware(ipLimiter))

	userH.RegisterRoutes(v1.Group("user"))
	relationH.RegisterRoutes(v1.Group("relation"))
	webhookH.RegisterRoutes(v1.Group("api"))

	return router
}

// buildSwaggerHandler returns a Gin handler that covers all /swagger/* paths.
//   - /swagger/ and /swagger/index.html → serve a custom Swagger UI page that
//     automatically reads eb_token from localStorage and pre-authorises the
//     ApiKeyAuth security scheme, so callers never need to paste a token manually.
//   - All other paths (JS / CSS / doc.json …) → delegated to the standard
//     ginSwagger handler so assets and the generated spec are served unchanged.
func buildSwaggerHandler() gin.HandlerFunc {
	swaggerIndex, err := templates.SwaggerUIHTML.ReadFile("swagger_ui.html")
	if err != nil {
		zap.S().Fatalf("failed to embed swagger_ui.html: %v", err)
	}
	assetHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)

	return func(c *gin.Context) {
		if p := c.Param("any"); p == "/" || p == "/index.html" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerIndex)
			return
		}
		assetHandler(c)
	}
}

