package config

import (
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// 全局配置实例
var (
	cfg  *Config
	once sync.Once
)

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	ClientID     string `mapstructure:"clientID"`
	ClientSecret string `mapstructure:"clientSecret"`
	EnableRedis  bool   `mapstructure:"enableRedis,default=false"`
}

type MailConfig struct {
	SmtpUser     string `mapstructure:"smtpUser"`
	SmtpPassword string `mapstructure:"smtpPassword"`
	MailFrom     string `mapstructure:"mailFrom"`
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
	Server          ServerConfig                      `mapstructure:"server"`
	MailConfig      MailConfig                        `mapstructure:"mail"`
	Databases       map[string]map[string]interface{} `mapstructure:"databases"`
	DefaultDatabase string                            `mapstructure:"default_database"`
	Redis           RedisConfig                       `mapstructure:"redis"`
	PgClient        ConnConfig                        `mapstructure:"pg_client"`
}

type ConnConfig struct {
	Addr         string        // for trace
	DSN          string        // write data source name.
	ReadDSN      []string      // read data source name.
	Active       int           // pool
	Idle         int           // pool
	IdleTimeout  time.Duration // connect max lifetime.
	QueryTimeout time.Duration // query sql timeout
	ExecTimeout  time.Duration // execute sql timeout
	TranTimeout  time.Duration // transaction sql timeout
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.AddConfigPath("config")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	v.AutomaticEnv() // 支持环境变量覆盖
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	zap.S().Info("✅ 配置加载成功")
	return cfg, nil
}
