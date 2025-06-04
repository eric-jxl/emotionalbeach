package initialize

import (
	"log"

	"go.uber.org/zap"
)

func InitLogger() {
	//初始化日志
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("日志初始化失败:%s", err.Error())
	}
	//使用全局logger
	defer logger.Sync()
	zap.ReplaceGlobals(logger)
	zap.S().Info("初始化Zap日志")
}
