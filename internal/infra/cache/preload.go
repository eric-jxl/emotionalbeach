// Package cache provides Redis warm-up helpers.
package cache

import (
	"context"
	"emotionalBeach/internal/models"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Preload warms up Redis with basic user data from the primary database.
// It is a no-op when rdb is nil (Redis disabled).
func Preload(rdb *redis.Client, db *gorm.DB) error {
	if rdb == nil {
		return nil
	}
	var users []struct {
		ID    int
		Name  string
		Role  string
		Email string
	}
	if err := db.Model(&models.UserBasic{}).Find(&users).Error; err != nil {
		return err
	}
	ctx := context.Background()
	for _, u := range users {
		key := fmt.Sprintf("user_%d", u.ID)
		data := map[string]interface{}{
			"name":  u.Name,
			"role":  u.Role,
			"email": u.Email,
		}
		if _, err := rdb.HSet(ctx, key, data).Result(); err != nil {
			return fmt.Errorf("redis HSet user_%d: %w", u.ID, err)
		}
	}
	return nil
}

