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

func InitDB(dbPath string) {
	errs := godotenv.Load(dbPath)
	if errs != nil {
		log.Fatalf("Error loading .env file: %v", errs)
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
	global.DB, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		zap.S().Error(err.Error())
	}
	dbErr := global.DB.AutoMigrate(&models.UserBasic{}, models.Relation{})
	zap.S().Info("数据库自动迁移...")
	if dbErr != nil {
		return
	}
	dbConnect, errs := global.DB.DB()
	if errs != nil {
		log.Fatalf("数据库错误%s\n", errs)
	}
	dbConnect.SetMaxIdleConns(10)           // 设置最大空闲连接数
	dbConnect.SetMaxOpenConns(100)          // 设置最大打开连接数
	dbConnect.SetConnMaxLifetime(time.Hour) // 设置连接最长生命周期（0表示没有限制）
}
