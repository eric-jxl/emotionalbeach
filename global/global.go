package global

import (
	"emotionalBeach/config"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
)

var (
	ServiceConfig *config.Config
	DB            *gorm.DB
	RedisClient   *redis.Client
)
