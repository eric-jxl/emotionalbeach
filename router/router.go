package router

import (
	"emotionalBeach/controller"
	_ "emotionalBeach/docs"
	"emotionalBeach/middlewear"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"path/filepath"
	"runtime"
)

func Router() *gin.Engine {
	router := gin.Default()
	//router := gin.New()
	//router.Use(middlewear.ZapLogger(), gin.Recovery())
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//router.GET("/", func(c *gin.Context) {
	//	c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	//})
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)

	// 拼接模板路径（../templates/*）
	templatePath := filepath.Join(basePath, "..", "templates", "*")
	router.LoadHTMLGlob(templatePath)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
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
	user := v1.Group("user").Use(middlewear.AuthJwt())

	{
		user.GET("/list", controller.GetUsers)
		user.GET("/condition", controller.GetAppointUser)
		user.DELETE("/delete", controller.DeleteUser)
		user.POST("/update", controller.UpdateUser)
	}

	//好友关系
	relation := v1.Group("relation").Use(middlewear.AuthJwt())
	{
		relation.POST("/list", controller.FriendList)
		relation.POST("/add", controller.AddFriendByName)
	}

	return router
}
