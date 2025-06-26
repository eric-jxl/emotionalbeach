package initialize

import (
	"log"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger() {
	//初始化日志
	var err error
	Logger, err = zap.NewDevelopment()
	if err != nil {
		log.Fatalf("日志初始化失败:%s", err.Error())
	}
	defer Logger.Sync()
	//使用全局logger
	zap.ReplaceGlobals(Logger)
	zap.S().Info("初始化Zap日志")
}
