package infra

import (
	"context"
	"emotionalBeach/config"
	"emotionalBeach/internal/models"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AutoMigrate runs GORM auto-migration when the -migrate / --migrate flag is
// present, then exits the process. No-op otherwise.
func AutoMigrate(db *gorm.DB) {
	for _, a := range os.Args {
		if a == "-migrate" || a == "--migrate" {
			if err := db.AutoMigrate(&models.UserBasic{}, &models.Relation{}); err != nil {
				zap.S().Fatalf("❌ migration failed: %v", err)
			}
			zap.S().Info("✅ migration done")
			os.Exit(0)
		}
	}
}

// CachePreload warms Redis with basic user data. No-op when Redis is disabled.
func CachePreload(cfg *config.Config, rdb *redis.Client, db *gorm.DB) {
	if !cfg.Server.EnableRedis || rdb == nil {
		return
	}
	if err := preloadRedis(rdb, db); err != nil {
		zap.S().Fatalf("❌ Redis preload failed: %v", err)
	}
	zap.S().Info("✅ Redis cache warmed up")
}

// RegisterCollectors registers the Prometheus DB-pool scrape-driven collector.
func RegisterCollectors(db *gorm.DB) {
	newDBPoolCollector(db)
}

func preloadRedis(rdb *redis.Client, db *gorm.DB) error {
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

