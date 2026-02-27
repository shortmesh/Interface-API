package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"interface-api/pkg/logger"

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

var dbInstance *service

type Options struct {
	AutoMigrate     bool
	AutoCreateTable bool
}

func New(opts ...Options) Service {
	if dbInstance != nil {
		return dbInstance
	}

	var options Options
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options.AutoMigrate = os.Getenv("AUTO_MIGRATE") == "true"
		options.AutoCreateTable = os.Getenv("AUTO_CREATE_TABLES") == "true"
	}

	dbname := os.Getenv("MYSQL_DATABASE")
	password := os.Getenv("MYSQL_PASSWORD")
	username := os.Getenv("MYSQL_USERNAME")
	port := os.Getenv("MYSQL_PORT")
	host := os.Getenv("MYSQL_HOST")
	sqlitePath := os.Getenv("SQLITE_DB_PATH")

	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		Logger: logger.NewGormLogger(),
	}

	if sqlitePath != "" {
		logger.Info(fmt.Sprintf("Using SQLite database: %s", sqlitePath))

		dir := filepath.Dir(sqlitePath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			logger.Error(fmt.Sprintf("Failed to create SQLite directory: %v", err))
			os.Exit(1)
		}

		db, err = gorm.Open(sqlite.Open(sqlitePath), gormConfig)
		if err != nil {
			logger.Error(fmt.Sprintf("SQLite connection failed: %v", err))
			os.Exit(1)
		}
	} else {
		logger.Info(fmt.Sprintf("Using MySQL database: %s@%s:%s/%s", username, host, port, dbname))
		loc, _ := time.LoadLocation("UTC")

		createDatabaseIfNotExists(username, password, host, port, dbname)

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
			username, password, host, port, dbname, loc)

		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			logger.Error(fmt.Sprintf("MySQL connection failed: %v", err))
			os.Exit(1)
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error(fmt.Sprintf("Database instance retrieval failed: %v", err))
		os.Exit(1)
	}

	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxIdleTime(90 * time.Second)

	dbInstance = &service{
		db: db,
	}

	if options.AutoMigrate {
		if err := dbInstance.AutoMigrate(options.AutoCreateTable); err != nil {
			logger.Error(fmt.Sprintf("Database migration failed: %v", err))
			os.Exit(1)
		}
	}

	return dbInstance
}

func createDatabaseIfNotExists(username, password, host, port, dbname string) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Error(fmt.Sprintf("MySQL server connection failed: %v", err))
		os.Exit(1)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	if err != nil {
		logger.Error(fmt.Sprintf("Database creation failed: %v", err))
		os.Exit(1)
	}
	logger.Info(fmt.Sprintf("Database %s ensured to exist", dbname))
}

func (s *service) DB() *gorm.DB {
	return s.db
}
