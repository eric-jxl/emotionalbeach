package global

import (
	"emotionalBeach/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	ServiceConfig *config.ServiceConfig
	DB            *gorm.DB
	RedisDB       *redis.Client
)
