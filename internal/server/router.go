package server

import (
	_ "emotionalBeach/docs"
	"emotionalBeach/internal/controller"
	"emotionalBeach/internal/middleware"
	"emotionalBeach/internal/service"
	"emotionalBeach/internal/templates"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), middleware.ZapLogger())
	// todo: ip 限流器
	ipLimiter := middleware.NewIPRateLimiter(rate.Every(10*time.Second), 5)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status_code": http.StatusNotFound,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"error":       "请求的资源不存在",
		})
	})
	// 健康检查
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 静态资源服务
	assetsFS, _ := fs.Sub(templates.AssetHTML, "assets")
	fileServer := http.FileServer(http.FS(assetsFS))
	router.GET("/assets/*filepath", func(c *gin.Context) {
		middleware.AssetsCacheMiddleware()(c)
		if c.IsAborted() {
			return
		}
		http.StripPrefix("/assets", fileServer).ServeHTTP(c.Writer, c.Request)
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/", func(c *gin.Context) {
		data, err := templates.IndexHTML.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error loading index.html")
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Github Login
	router.GET("/login/github", controller.GithubLogin)
	router.GET("/callback", controller.GithubCallback)

	router.Any("/login", func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodPost:
			controller.LoginByNameAndPassWord(c)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"code": http.StatusMethodNotAllowed, "message": "Only  POST methods are allowed"})
		}

	})
	router.POST("/register", controller.NewUser)

	// TODO: v1版本接口
	v1 := router.Group("/v1", middleware.AuthJwt(), middleware.RateLimitMiddleware(ipLimiter))
	//用户接口
	user := v1.Group("user")

	{
		user.GET("/list", controller.GetUsers)
		user.GET("/condition", controller.GetAppointUser)
		user.DELETE("/delete", controller.DeleteUser)
		user.POST("/update", controller.UpdateUser)
	}

	//好友关系
	relation := v1.Group("relation")
	{
		relation.POST("/list", controller.FriendList)
		relation.POST("/add", controller.AddFriendByName)
	}
	apiV1 := v1.Group("/api")
	{
		apiV1.POST("/webhook", service.WebhookEmail)
	}

	return router
}
