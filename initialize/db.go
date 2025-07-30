package initialize

import (
	"emotionalBeach/global"
	"emotionalBeach/models"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(dbPath string) error {
	errs := godotenv.Load(dbPath)
	if errs != nil {
		log.Fatalf("Error loading .env file: %v", errs)
		return errs
	}
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", dbHost, dbUser, dbPassword, dbName, dbPort)
	//写sql语句配置
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        // 彩色打印
		},
	)

	//将获取到的连接赋值到global.DB
	config := &gorm.Config{}
	if gin.Mode() == gin.DebugMode {
		config.Logger = newLogger
	}
	var err error
	if global.DB, err = gorm.Open(postgres.Open(dsn), config); err != nil {
		zap.S().Error(err.Error())
		return err
	}
	if sqlDB, tmpErr := global.DB.DB(); tmpErr == nil {
		sqlDB.SetMaxOpenConns(20)           //  设置最大打开连接数
		sqlDB.SetMaxIdleConns(100)          // 设置最大空闲连接数
		sqlDB.SetConnMaxLifetime(time.Hour) // 设置连接最长生命周期（0表示没有限制）
		sqlDB.SetConnMaxIdleTime(time.Hour * 1)
	}
	zap.S().Info("数据库自动迁移中...")
	dbErr := global.DB.AutoMigrate(&models.UserBasic{}, models.Relation{})
	zap.S().Info("数据库自动迁移完成")
	if dbErr != nil {
		return dbErr
	}
	return nil
}
