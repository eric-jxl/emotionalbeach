package controller

import (
	"context"
	"emotionalBeach/internal/models"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func PreloadCache(redisClient *redis.Client, db *gorm.DB) error {
	// 查询数据
	var users []struct {
		ID    int
		Name  string
		Role  string
		Email string
	}
	if err := db.Model(&models.UserBasic{}).Find(&users).Error; err != nil {
		return err
	}
	// 写入 Redis
	for _, p := range users {
		key := fmt.Sprintf("user_%d", p.ID)
		data := map[string]interface{}{
			"name":  p.Name,
			"role":  p.Role,
			"email": p.Email,
		}
		_, err := redisClient.HSet(context.Background(), key, data).Result()
		if err != nil {
			return err
		}
	}

	return nil
}
