package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"interface-api/internal/logger"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Service interface {
	DB() *gorm.DB
}

type service struct {
	db *gorm.DB
}

var (
	dbname     = os.Getenv("MYSQL_DATABASE")
	password   = os.Getenv("MYSQL_PASSWORD")
	username   = os.Getenv("MYSQL_USERNAME")
	port       = os.Getenv("MYSQL_PORT")
	host       = os.Getenv("MYSQL_HOST")
	sqlitePath = os.Getenv("SQLITE_DB_PATH")
	dbInstance *service
)

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	var db *gorm.DB
	var err error

	if sqlitePath != "" {
		logger.Log.Infof("Using SQLite database: %s", sqlitePath)
		db, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
		if err != nil {
			logger.Log.Fatalf("Failed to connect to SQLite database: %v", err)
		}
	} else {
		logger.Log.Infof("Using MySQL database: %s@%s:%s/%s", username, host, port, dbname)
		loc, _ := time.LoadLocation("UTC")

		createDatabaseIfNotExists(username, password, host, port, dbname)

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
			username, password, host, port, dbname, loc)

		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			logger.Log.Fatalf("Failed to connect to MySQL database: %v", err)
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Log.Fatalf("Failed to get database instance: %v", err)
	}

	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(50)

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func createDatabaseIfNotExists(username, password, host, port, dbname string) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Log.Fatalf("Failed to connect to MySQL server: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	if err != nil {
		logger.Log.Fatalf("Failed to create database: %v", err)
	}
	logger.Log.Infof("Database %s ensured to exist", dbname)
}

func (s *service) DB() *gorm.DB {
	return s.db
}
