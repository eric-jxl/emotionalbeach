package main

import (
	"emotionalBeach/config"
	"emotionalBeach/initialize"
	"emotionalBeach/router"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

//var (
//	filepath *string
//)
//
//func init() {
//	filepath = flag.String("e", "env", "数据库配置文件(.env)路径")
//	flag.Parse()
//}

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
	//初始化日志
	initialize.InitLogger()
	//初始化数据库
	cfg, dbErr := config.LoadConfig()
	if dbErr != nil {
		log.Fatalf("❌ 加载配置失败: %v", dbErr)
	}

	// 初始化数据库
	if err := initialize.InitDatabases(cfg.Databases, cfg.Database.Default); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	log.Printf("✅ 数据库连接成功")

	if _, err := initialize.InitRedis(cfg.Redis); err != nil {
		log.Fatalf("❌ Redis 初始化失败: %v", err)
	}
	log.Println("✅ Redis 连接成功")
	initialize.StartDatabases()
	//exception := initialize.InitDB(*filepath)
	//if exception != nil {
	//	panic(exception)
	//}
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
	zap.S().Info("Initiating shutdown")
	zap.S().Info("Hit CTRL-C again or send a second signal to force the shutdown.")
}
