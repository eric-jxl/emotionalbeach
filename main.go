package main

import (
	"emotionalBeach/config"
	"emotionalBeach/internal/di"
	"log"

	"go.uber.org/zap"
)

// @title           情感沙滩 API
// @version         1.0
// @description     Development Environment: go v1.23.7 + gin v1.10.1 + gorm v1.30.2 + viper v1.20.1
// @contact.name    Eric Jiang
// @contact.url     http://www.swagger.io/support
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath        /
// @securityDefinitions.apikey ApiKeyAuth
// @type            apiKey
// @in              header
// @name            Authorization

//go:generate swag init -o ./docs -g main.go
func main() {
	bootstrapLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ load config failed: %v", err)
	}
	zap.S().Infof("✅ 服务运行在端口: \x1b[32m%d\x1b[0m", cfg.Server.Port)

	// initialises DB/Redis, builds the HTTP server, and returns a cleanup func.
	application, cleanup, err := di.InitializeApp(cfg)
	if err != nil {
		zap.S().Fatalf("❌ initialize app failed: %v", err)
	}
	defer cleanup()

	application.Run()
}

// bootstrapLogger sets up a minimal zap console logger so that log calls
func bootstrapLogger() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}
