package infra

import (
	"context"
	"emotionalBeach/config"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var gormLog = gormlogger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags),
	gormlogger.Config{
		SlowThreshold:             2 * time.Second,
		LogLevel:                  gormlogger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	},
)

// ProvideDB initialises all configured databases and returns the main *gorm.DB.
func ProvideDB(cfg *config.Config) (*gorm.DB, func(), error) {
	dbs := make(map[string]*gorm.DB)

	for name, raw := range cfg.Databases {
		typ, ok := raw["type"].(string)
		if !ok {
			return nil, nil, fmt.Errorf("database %q missing 'type' field", name)
		}
		var (
			gdb *gorm.DB
			err error
		)
		switch typ {
		case "postgres":
			var pg config.PostgresConfig
			dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				WeaklyTypedInput: true, TagName: "mapstructure", Result: &pg,
			})
			if err = dec.Decode(raw); err != nil {
				return nil, nil, fmt.Errorf("decode postgres config for %q: %w", name, err)
			}
			gdb, err = openPostgres(pg)
		case "mysql":
			var mc config.MySQLConfig
			dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				WeaklyTypedInput: true, TagName: "mapstructure", Result: &mc,
			})
			if err = dec.Decode(raw); err != nil {
				return nil, nil, fmt.Errorf("decode mysql config for %q: %w", name, err)
			}
			gdb, err = openMySQL(mc)
		default:
			zap.S().Warnf("⚠️  unknown database type %q, skipping", typ)
			continue
		}
		if err != nil {
			return nil, nil, fmt.Errorf("open database %q: %w", name, err)
		}
		dbs[name] = gdb
		zap.S().Infof("✅ database [%s] (%s) connected", name, typ)
	}

	mainDB, ok := dbs[cfg.DefaultDatabase]
	if !ok {
		return nil, nil, fmt.Errorf("default database %q not found in config", cfg.DefaultDatabase)
	}
	zap.S().Infof("🎯 default DB: %s", cfg.DefaultDatabase)

	cleanup := func() {
		for name, db := range dbs {
			sqlDB, err := db.DB()
			if err != nil {
				zap.S().Warnf("skip closing db %s: %v", name, err)
				continue
			}
			if err = sqlDB.Close(); err != nil {
				zap.S().Warnf("close db %s: %v", name, err)
			}
		}
		zap.S().Info("🗄️  all databases closed")
	}
	return mainDB, cleanup, nil
}

// ProvideRedis initialises a Redis client when EnableRedis is true.
// Returns a nil client (no error) when Redis is disabled.
func ProvideRedis(cfg *config.Config) (*redis.Client, func(), error) {
	if !cfg.Server.EnableRedis {
		return nil, func() {}, nil
	}
	c := cfg.Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password:     c.Password,
		DB:           c.DB,
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConns,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("redis ping: %w", err)
	}
	zap.S().Info("✅ Redis connected")
	cleanup := func() {
		if err := rdb.Close(); err != nil {
			zap.S().Warnf("close redis: %v", err)
		}
		zap.S().Info("🗄️  Redis closed")
	}
	return rdb, cleanup, nil
}

// dbSet is internal — the wire.ProviderSet for DB providers (used by Provider).
var dbSet = wire.NewSet(ProvideDB, ProvideRedis)

func openPostgres(cfg config.PostgresConfig) (*gorm.DB, error) {
	if cfg.Host == "" || cfg.User == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("postgres config missing host/user/dbname")
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode,
	)
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLog})
	if err != nil {
		return nil, err
	}
	return applyPool(gdb, cfg.DBCommon)
}

func openMySQL(cfg config.MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
		cfg.Charset, cfg.ParseTime, cfg.Loc)
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: gormLog})
	if err != nil {
		return nil, err
	}
	return applyPool(gdb, cfg.DBCommon)
}

func applyPool(gdb *gorm.DB, cfg config.DBCommon) (*gorm.DB, error) {
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.SetMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Second)
	if err = sqlDB.Ping(); err != nil {
		return nil, err
	}
	return gdb, nil
}

