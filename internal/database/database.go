package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"interface-api/pkg/logger"

	sqlcipher "github.com/gdanko/gorm-sqlcipher"
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

	logger.Info("Initializing encrypted SQLite database with SQLCipher")
	dsn := fmt.Sprintf(
		"%s?_pragma_key=x'%s'&_pragma_cipher_compatibility=3&_pragma_cipher_page_size=4096",
		sqlitePath,
		dbEncryptionKey,
	)
	db, err = gorm.Open(sqlcipher.Open(dsn), gormConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Encrypted SQLite connection failed: %v", err))
		os.Exit(1)
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
