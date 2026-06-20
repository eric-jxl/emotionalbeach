package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	ClientID     string `mapstructure:"clientID"`
	ClientSecret string `mapstructure:"clientSecret"`
	JWTSecret    string `mapstructure:"jwtSecret"`
	// ESA captcha runtime params are served by backend endpoint to avoid hardcoding in static HTML.
	ESACaptchaRegion  string `mapstructure:"esaCaptchaRegion"`
	ESACaptchaPrefix  string `mapstructure:"esaCaptchaPrefix"`
	ESACaptchaSceneID string `mapstructure:"esaCaptchaSceneID"`
	// Rate-limit policy for all protected /v1 routes.
	RateLimitWindowSec int  `mapstructure:"rateLimitWindowSec"`
	RateLimitMaxReq    int  `mapstructure:"rateLimitMaxReq"`
	EnableRedis        bool `mapstructure:"enableRedis,default=false"`
	ReadTimeoutSec     int  `mapstructure:"readTimeoutSec"`
	WriteTimeoutSec    int  `mapstructure:"writeTimeoutSec"`
	IdleTimeoutSec     int  `mapstructure:"idleTimeoutSec"`
	ShutdownTimeoutSec int  `mapstructure:"shutdownTimeoutSec"`
}

func (s ServerConfig) ReadTimeout() time.Duration {
	return time.Duration(s.ReadTimeoutSec) * time.Second
}

func (s ServerConfig) WriteTimeout() time.Duration {
	return time.Duration(s.WriteTimeoutSec) * time.Second
}

func (s ServerConfig) IdleTimeout() time.Duration {
	return time.Duration(s.IdleTimeoutSec) * time.Second
}

func (s ServerConfig) ShutdownTimeout() time.Duration {
	return time.Duration(s.ShutdownTimeoutSec) * time.Second
}

func (s ServerConfig) RateLimitEvery() time.Duration {
	return time.Duration(s.RateLimitWindowSec) * time.Second
}

func (s ServerConfig) RateLimitBurst() int {
	return s.RateLimitMaxReq
}

func (s ServerConfig) CaptchaRegion() string {
	if s.ESACaptchaRegion == "" {
		return "cn"
	}
	return s.ESACaptchaRegion
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

// LogConfig 日志滚动与级别配置
type LogConfig struct {
	Level      string `mapstructure:"level"`      // debug | info | warn | error
	Filename   string `mapstructure:"filename"`   // 留空则只输出控制台
	MaxSizeMB  int    `mapstructure:"maxSizeMB"`  // 单个日志文件最大 MB
	MaxBackups int    `mapstructure:"maxBackups"` // 保留旧日志文件个数
	MaxAgeDays int    `mapstructure:"maxAgeDays"` // 日志保留天数
	Compress   bool   `mapstructure:"compress"`   // 旧日志是否 gzip 压缩
}

type Config struct {
	Server          ServerConfig                      `mapstructure:"server"`
	Log             LogConfig                         `mapstructure:"log"`
	MailConfig      MailConfig                        `mapstructure:"mail"`
	Databases       map[string]map[string]interface{} `mapstructure:"databases"`
	DefaultDatabase string                            `mapstructure:"default_database"`
	Redis           RedisConfig                       `mapstructure:"redis"`
}

var defaultValues = map[string]any{
	// Security- and availability-related defaults live in one place for easier review.
	"server.readTimeoutSec":     15,
	"server.writeTimeoutSec":    30,
	"server.idleTimeoutSec":     60,
	"server.shutdownTimeoutSec": 10,
	"server.esaCaptchaRegion":   "cn",
	"server.rateLimitWindowSec": 10,
	"server.rateLimitMaxReq":    5,
	"log.level":                 "info",
	"log.maxSizeMB":             100,
	"log.maxBackups":            5,
	"log.maxAgeDays":            30,
	"log.compress":              true,
}

func applyDefaults(v *viper.Viper) {
	for key, value := range defaultValues {
		v.SetDefault(key, value)
	}
}

// normalizeCompatKeys preserves backward compatibility for renamed config keys.
func normalizeCompatKeys(v *viper.Viper) {
	if !v.IsSet("redis.min_idle_conns") && v.IsSet("redis.min_idle_connects") {
		v.Set("redis.min_idle_conns", v.Get("redis.min_idle_connects"))
	}
}

func LoadConfig() (cfg *Config, err error) {
	v := viper.New()
	v.AddConfigPath("config")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	applyDefaults(v)
	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv() // 支持环境变量覆盖
	normalizeCompatKeys(v)
	if err = v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	zap.S().Info("✅ 配置加载成功")
	return cfg, nil
}
