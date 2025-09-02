package initialize

import (
	"context"
	"emotionalBeach/config"
	"emotionalBeach/global"
	"emotionalBeach/models"
	"fmt"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
)

var (
	DBs    = map[string]*gorm.DB{}
	MainDB *gorm.DB
)

// InitDatabases 初始化所有数据库连接
func InitDatabases(cfg map[string]map[string]interface{}, defaultDB string) error {
	for name, raw := range cfg {
		typ, ok := raw["type"].(string)
		if !ok {
			return fmt.Errorf("数据库 %s 缺少 type 配置", name)
		}

		var gdb *gorm.DB
		var err error

		switch typ {
		case "postgres":
			pg := config.PostgresConfig{}
			decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				WeaklyTypedInput: true, // ⭐ 自动类型转换 float64->int, string->time.Duration
				TagName:          "mapstructure",
				Result:           &pg,
			})
			if err := decoder.Decode(raw); err != nil {
				return fmt.Errorf("数据库 %s 配置解析失败: %v", name, err)
			}
			gdb, err = initPostgres(pg)

		case "mysql":
			mysqlCfg := config.MySQLConfig{}
			decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
				TagName:          "mapstructure",
				Result:           &mysqlCfg,
			})
			if err := decoder.Decode(raw); err != nil {
				return fmt.Errorf("数据库 %s 配置解析失败: %v", name, err)
			}
			gdb, err = initMySQL(mysqlCfg)

		default:
			zap.S().Warnf("⚠️ 未知数据库类型: %s (跳过)", typ)
			continue
		}

		if err != nil {
			return fmt.Errorf("初始化数据库 %s 失败: %v", name, err)
		}

		DBs[name] = gdb
		zap.S().Infof("✅ 数据库 [%s] (%s) 连接成功", name, typ)
	}

	// 设置默认数据库
	if db, ok := DBs[defaultDB]; ok {
		MainDB = db
		zap.S().Infof("🎯 默认 ORM 数据库: %s", defaultDB)
	} else {
		return fmt.Errorf("指定的默认数据库 [%s] 不存在", defaultDB)
	}

	return nil
}

func initPostgres(cfg config.PostgresConfig) (*gorm.DB, error) {
	// 校验必填字段
	if cfg.Host == "" || cfg.User == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("postgres 配置缺少必要字段 (host:%s/user:%s/dbname%s)", cfg.Host, cfg.User, cfg.DBName)
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai", cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	return setupPool(gdb, cfg.DBCommon)
}

func initMySQL(cfg config.MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
		cfg.Charset, cfg.ParseTime, cfg.Loc)

	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	return setupPool(gdb, cfg.DBCommon)
}

func setupPool(gdb *gorm.DB, cfg config.DBCommon) (*gorm.DB, error) {
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.SetMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	return gdb, nil
}

// Get 获取指定数据库
func Get(name string) (*gorm.DB, bool) {
	db, ok := DBs[name]
	return db, ok
}

// GetDefault 获取默认数据库
func GetDefault() *gorm.DB {
	return MainDB
}

// InitRedis 初始化Redis
func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	global.RedisClient = rdb
	zap.S().Info("✅ Redis 连接成功")
	return rdb, nil
}

// StartDatabases 自动迁移数据库
func StartDatabases(config *config.Config) (err error) {
	// 初始化数据库
	if err = InitDatabases(config.Databases, config.Database.Default); err != nil {
		zap.S().Fatalf("❌ 数据库初始化失败: %v", err)
		return
	}

	if _, err = InitRedis(config.Redis); err != nil {
		zap.S().Fatalf("❌ Redis 初始化失败: %v", err)
		return
	}
	err = GetDefault().AutoMigrate(&models.UserBasic{}, &models.Relation{})
	if err != nil {
		zap.S().Fatalf("启动出错: %v", err)
		return
	}
	return
}
