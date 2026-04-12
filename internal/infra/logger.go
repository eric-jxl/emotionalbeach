package infra

import (
	"emotionalBeach/config"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/wire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Loggers bundles the system logger and the HTTP access logger.
type Loggers struct {
	Sys    *zap.Logger // structured, with caller + stacktrace on errors
	Access *zap.Logger // HTTP access log (console: message-only; file: full JSON)
}

// messageOnlyCore wraps a Core and drops all fields before writing.
type messageOnlyCore struct{ zapcore.Core }

func (m messageOnlyCore) With(_ []zapcore.Field) zapcore.Core { return m }
func (m messageOnlyCore) Write(e zapcore.Entry, _ []zapcore.Field) error {
	return m.Core.Write(e, nil)
}
func (m messageOnlyCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return m.Core.Check(e, ce)
}

// ProvideLoggers builds both loggers, replaces the zap global, and returns a cleanup func.
func ProvideLoggers(cfg *config.Config) (*Loggers, func(), error) {
	lcfg := cfg.Log
	level := parseLevel(lcfg.Level)

	consoleCfg := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleCfg),
		zapcore.Lock(os.Stdout),
		level,
	)

	var fileCore zapcore.Core
	if lcfg.Filename != "" {
		_ = os.MkdirAll(filepath.Dir(lcfg.Filename), 0755)
		rot := &lumberjack.Logger{
			Filename:   lcfg.Filename,
			MaxSize:    lcfg.MaxSizeMB,
			MaxBackups: lcfg.MaxBackups,
			MaxAge:     lcfg.MaxAgeDays,
			Compress:   lcfg.Compress,
		}
		fileCfg := zap.NewProductionEncoderConfig()
		fileCfg.TimeKey = "time"
		fileCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		fileCore = zapcore.NewCore(zapcore.NewJSONEncoder(fileCfg), zapcore.AddSync(rot), level)
	}

	sysCore := zapcore.Core(consoleCore)
	if fileCore != nil {
		sysCore = zapcore.NewTee(consoleCore, fileCore)
	}
	sys := zap.New(sysCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(sys)

	accessConsole := messageOnlyCore{consoleCore}
	var accessCore zapcore.Core = accessConsole
	if fileCore != nil {
		accessCore = zapcore.NewTee(accessConsole, fileCore)
	}
	access := zap.New(accessCore)
	zap.S().Info("🌏️  logger initialized")

	cleanup := func() {
		_ = sys.Sync()
		_ = access.Sync()
	}
	return &Loggers{Sys: sys, Access: access}, cleanup, nil
}

// loggerSet is internal — used by Provider.
var loggerSet = wire.NewSet(ProvideLoggers)

func parseLevel(s string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

