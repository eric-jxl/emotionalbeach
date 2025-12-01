package server

import (
	_ "emotionalBeach/docs"
	"emotionalBeach/internal/controller"
	"emotionalBeach/internal/middleware"
	"emotionalBeach/internal/service"
	"emotionalBeach/internal/templates"
	"io/fs"
	"net/http"
	"text/template"
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

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error":            "请求的方法不被允许",
			"status_code":      http.StatusMethodNotAllowed,
			"allowed_methods":  c.GetHeader("Allow"), // Gin 通常会在 Allow 头部列出允许的方法
			"requested_method": c.Request.Method,
			"requested_path":   c.Request.URL.Path,
		})
	})
	// 静态文件服务
	fsys, err := fs.Sub(templates.AssetHTML, "assets")
	if err != nil {
		panic(err)
	}
	// 静态资源服务
	router.StaticFS("/assets", http.FS(fsys))
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

	router.GET("/dir", func(c *gin.Context) {
		files, err := templates.IndexHTML.ReadDir(".")
		if err != nil {
			c.String(http.StatusInternalServerError, "无法读取目录")
			return
		}

		var fileList []string
		for _, file := range files {
			fileList = append(fileList, file.Name())
		}
		t := template.New("Dir")
		t = template.Must(t.Parse(templates.DirHTMLContent))
		err = t.Execute(c.Writer, fileList)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "模板渲染失败"})
		}
	})
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.Any("/login", func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodPost:
			controller.LoginByNameAndPassWord(c)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"code": http.StatusMethodNotAllowed, "message": "Only  POST methods are allowed"})
		}

	})
	router.POST("/register", controller.NewUser)

	v1 := router.Group("/v1")
	//用户接口
	user := v1.Group("user").Use(middleware.AuthJwt()).Use(middleware.RateLimitMiddleware(ipLimiter))

	{
		user.GET("/list", controller.GetUsers)
		user.GET("/condition", controller.GetAppointUser)
		user.DELETE("/delete", controller.DeleteUser)
		user.POST("/update", controller.UpdateUser)
	}

	//好友关系
	relation := v1.Group("relation").Use(middleware.AuthJwt())
	{
		relation.POST("/list", controller.FriendList)
		relation.POST("/add", controller.AddFriendByName)
	}
	apiV1 := v1.Group("/api", middleware.AuthJwt()).Use(middleware.RateLimitMiddleware(ipLimiter))
	{
		apiV1.POST("/webhook", service.WebhookEmail)
	}

	return router
}
