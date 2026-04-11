// Package logger provides Wire-compatible providers for structured logging.
package logger

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

// Set is the Wire provider set for logging.
var Set = wire.NewSet(ProvideLoggers)

// Loggers bundles the system logger and the HTTP access logger.
// Keeping them in a single struct avoids Wire's ambiguity around
// two *zap.Logger values in the same dependency graph.
type Loggers struct {
	Sys    *zap.Logger // structured, with caller + stacktrace on errors
	Access *zap.Logger // HTTP access log (console: message-only; file: full JSON)
}

// messageOnlyCore wraps a Core and drops all fields before writing.
// Used for the console access-log so we get a clean single-line format
// instead of the noisy "key=value" appended by default.
type messageOnlyCore struct{ zapcore.Core }

func (m messageOnlyCore) With(_ []zapcore.Field) zapcore.Core { return m }
func (m messageOnlyCore) Write(e zapcore.Entry, _ []zapcore.Field) error {
	return m.Core.Write(e, nil)
}
func (m messageOnlyCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return m.Core.Check(e, ce)
}

// ProvideLoggers builds both loggers from configuration, replaces the zap
// global logger, and returns a cleanup func that syncs both writers.
func ProvideLoggers(cfg *config.Config) (*Loggers, func(), error) {
	lcfg := cfg.Log
	level := parseLevel(lcfg.Level)

	// ── Console core: human-readable with colour ──────────────────────────
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

	// ── File core: JSON structured with lumberjack rotation ───────────────
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

	// ── System logger: console + file, with caller & error stacktrace ─────
	sysCore := zapcore.Core(consoleCore)
	if fileCore != nil {
		sysCore = zapcore.NewTee(consoleCore, fileCore)
	}
	sys := zap.New(sysCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(sys)

	// ── Access logger: console (message-only) + file (full JSON) ─────────
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

