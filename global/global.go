package global

import (
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
)

var (
	DB          *gorm.DB
	RedisClient *redis.Client
)
