package config

import (
	"time"

	"go.uber.org/zap"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	ClientID     string `mapstructure:"clientID"`
	ClientSecret string `mapstructure:"clientSecret"`
	EnableRedis  bool   `mapstructure:"enableRedis,default=false"`
}

type DBCommon struct {
	Type            string        `mapstructure:"type"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	SetMaxIdleConns int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// PostgresConfig  专有字段
type PostgresConfig struct {
	DBCommon `mapstructure:",squash"`
	SSLMode  string `mapstructure:"sslmode"`
}

// MySQLConfig  专有字段
type MySQLConfig struct {
	DBCommon  `mapstructure:",squash"`
	Charset   string `mapstructure:"charset"`
	ParseTime bool   `mapstructure:"parse_time"`
	Loc       string `mapstructure:"loc"`
}
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type Config struct {
	Server    ServerConfig                      `mapstructure:"server"`
	Databases map[string]map[string]interface{} `mapstructure:"databases"`
	Database  struct {
		Default string `mapstructure:"default"`
	} `mapstructure:"database"`
	Redis RedisConfig `mapstructure:"redis"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	v.WatchConfig()
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	zap.S().Info("✅ 配置加载成功")
	return &cfg, nil
}
