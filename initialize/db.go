package initialize

import (
	"emotionalBeach/global"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() {
	errs := godotenv.Load("../config/.env")
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
			Colorful:                  true,        // 禁用彩色打印
		},
	)

	var err error
	//将获取到的连接赋值到global.DB
	global.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger, //打印sql日志
	})
	if err != nil {
		panic(err)
	}
	global.DB.SetMaxIdleConns(10) // 设置最大空闲连接数
	sqlDB.SetMaxOpenConns(100)    // 设置最大打开连接数
	sqlDB.SetConnMaxLifetime(0)   // 设置连接最长生命周期（0表示没有限制）
}
