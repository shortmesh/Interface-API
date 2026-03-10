package database

import (
	"fmt"

	"interface-api/migrations"
	"interface-api/pkg/logger"
	"interface-api/pkg/migrator"
)

func (s *service) RunMigrations() error {
	logger.Info("Running database migrations")

	scripts := migrations.GetAllMigrations()
	manager := migrator.NewManager(s.db, scripts)

	if err := manager.Up(); err != nil {
		logger.Error(fmt.Sprintf("Migration execution failed: %v", err))
		return err
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

func (s *service) AutoMigrate() error {
	return s.RunMigrations()
}
