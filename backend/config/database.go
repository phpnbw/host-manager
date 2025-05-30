package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	var err error

	// 从环境变量获取数据库类型，默认为sqlite
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	switch dbType {
	case "mysql":
		err = initMySQL()
	case "sqlite":
		err = initSQLite()
	default:
		log.Fatal("Unsupported database type:", dbType)
	}

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Printf("Database connected successfully (%s)\n", dbType)
}

func initSQLite() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "host_manager.db"
	}

	var err error
	// 使用 SQLite 驱动，在 Alpine Linux 中可能需要特殊处理
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	return err
}

func initMySQL() error {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "123456"
	}

	database := os.Getenv("DB_NAME")
	if database == "" {
		database = "host_manager"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, database)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return err
}
