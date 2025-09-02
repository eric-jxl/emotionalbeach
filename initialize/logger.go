package initialize

import (
	"os"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger() {
	//初始化日志
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,                   // 日志级别加颜色
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"), // 设置日期时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
	// 日志输出目标，控制台输出
	core := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)

	Logger = zap.New(core)
	defer Logger.Sync()
	//使用全局logger
	zap.ReplaceGlobals(Logger)
	zap.S().Info("初始化Zap日志")
}
