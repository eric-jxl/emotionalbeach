package main

import (
	"emotionalBeach/config"
	"emotionalBeach/controller"
	"emotionalBeach/global"
	"emotionalBeach/initialize"
	"emotionalBeach/router"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// @title 情感沙滩API
// @version 1.0
// @description ```
// @description Development Environment :go v1.23.7 + gin v1.10.1 + gorm v1.30.2 + viper v1.20.1
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
	//初始化日志
	initialize.InitLogger()
	//初始化数据库
	cfg, dbErr := config.LoadConfig()
	if dbErr != nil {
		zap.S().Fatalf("❌ 加载配置失败: %v", dbErr)
	}
	zap.S().Infof("Server run port : %d", cfg.Server.Port)

	dbErrs := initialize.StartDatabases(cfg)
	if dbErrs != nil {
		zap.S().Fatalf("启动数据库和Redis失败: %v", dbErrs.Error())
	}
	rdErr := controller.PreloadCache(global.RedisClient, initialize.MainDB)
	if rdErr != nil {
		zap.S().Fatalf("Redis 预热失败: %v", rdErr.Error())
	}
	// 启动服务
	routers := router.Router()
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	go func() {
		if err := routers.Run(port); err != nil {
			zap.S().Error(err.Error())
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Warn("Initiating shutdown")
	zap.S().Warn("Hit CTRL-C again or send a second signal to force the shutdown.")
}
