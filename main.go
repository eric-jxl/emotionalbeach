package main

import (
	"emotionalBeach/initialize"
	"emotionalBeach/router"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	filepath *string
)

func init() {
	filepath = flag.String("e", "env", "数据库配置文件(.env)路径")
	flag.Parse()
}

// @title 情感沙滩API
// @version 1.0
// @description ```
// @description Development Environment :go v1.23.7 + gin v1.10.1 + gorm v1.30.0
// @description ```

// @contact.name Eric Jiang
// @contact.url http://www.swagger.io/support
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @type apiKey
// @in header
// @name Authorization

//go:generate swag init -o ./docs -g main.go
func main() {
	gin.SetMode(gin.ReleaseMode)
	//初始化日志
	initialize.InitLogger()
	//初始化数据库
	exception := initialize.InitDB(*filepath)
	if exception != nil {
		panic(exception)
	}
	routers := router.Router()
	go func() {
		if err := routers.Run(":8080"); err != nil {
			zap.S().Error(err.Error())
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Info("Initiating shutdown")
	zap.S().Info("Hit CTRL-C again or send a second signal to force the shutdown.")
}
