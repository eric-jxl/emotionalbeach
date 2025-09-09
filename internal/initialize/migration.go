package initialize

import (
	"emotionalBeach/internal/models"
	"flag"
	"os"

	"go.uber.org/zap"
)

var migrate *bool

func init() {
	migrate = flag.Bool("migrate", false, "run database migration")
	flag.Parse()
}

func Migrate() {
	if *migrate {
		errAutoMigrate := MainDB.AutoMigrate(&models.UserBasic{}, &models.Relation{})
		if errAutoMigrate != nil {
			zap.S().Errorf("数据库迁移错误: %v", errAutoMigrate.Error())
		}
		zap.S().Info("Migration done ✅")
		os.Exit(0)
	}
}
