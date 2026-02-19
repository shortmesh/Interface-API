package database

import (
	"os"

	"interface-api/internal/database/models"
	"interface-api/migrations"
	"interface-api/pkg/logger"
	"interface-api/pkg/migrator"

	_ "github.com/joho/godotenv/autoload"
)

func (s *service) CreateTables() error {
	logger.Log.Info("Creating database tables")

	err := s.db.AutoMigrate(
		&models.User{}, &models.Session{}, &models.MatrixProfile{},
	)
	if err != nil {
		logger.Log.Errorf("Table creation failed: %v", err)
		return err
	}

	logger.Log.Info("Database tables created successfully")
	return nil
}

func (s *service) RunMigrations() error {
	logger.Log.Info("Running database migrations")

	scripts := migrations.GetAllMigrations()
	manager := migrator.NewManager(s.db, scripts)

	if err := manager.Up(); err != nil {
		logger.Log.Errorf("Migration execution failed: %v", err)
		return err
	}

	logger.Log.Info("Database migrations completed successfully")
	return nil
}

func (s *service) AutoMigrate() error {
	autoCreateTables := os.Getenv("AUTO_CREATE_TABLES")
	if autoCreateTables == "true" {
		if err := s.CreateTables(); err != nil {
			return err
		}
	}

	return s.RunMigrations()
}
