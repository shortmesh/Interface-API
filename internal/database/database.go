package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"interface-api/pkg/config"
	"interface-api/pkg/logger"

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
	AutoMigrate bool
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
	}

	sqlitePath := os.Getenv("SQLITE_DB_PATH")
	if sqlitePath == "" {
		sqlitePath = "./data/shortmesh.db"
	}
	dbEncryptionKey := os.Getenv("DB_ENCRYPTION_KEY")
	disableEncryption := strings.ToLower(os.Getenv("DISABLE_DB_ENCRYPTION"))

	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		Logger: logger.NewGormLogger(),
	}

	logger.Info(fmt.Sprintf("Using SQLite database: %s", sqlitePath))

	dir := filepath.Dir(sqlitePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		logger.Error(fmt.Sprintf("Failed to create SQLite directory: %v", err))
		os.Exit(1)
	}

	shouldEncrypt := true
	if disableEncryption == "true" {
		shouldEncrypt = false
		if config.IsProd() {
			logger.Warn("SECURITY WARNING: Database encryption disabled in production mode")
		}
	}

	if shouldEncrypt {
		logger.Info("Initializing encrypted SQLite database with SQLCipher")
		db, err = openDatabase(sqlitePath, dbEncryptionKey, gormConfig)
		if err != nil {
			logger.Error(fmt.Sprintf("Encrypted SQLite connection failed: %v", err))
			os.Exit(1)
		}
	} else {
		logger.Info("Initializing standard SQLite database (unencrypted)")
		db, err = openDatabase(sqlitePath, "", gormConfig)
		if err != nil {
			logger.Error(fmt.Sprintf("SQLite connection failed: %v", err))
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
		if err := dbInstance.AutoMigrate(); err != nil {
			logger.Error(fmt.Sprintf("Database migration failed: %v", err))
			os.Exit(1)
		}
	}

	return dbInstance
}

func (s *service) DB() *gorm.DB {
	return s.db
}
