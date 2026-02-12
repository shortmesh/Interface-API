package migrator

import (
	"errors"
	"fmt"

	"interface-api/internal/logger"

	"gorm.io/gorm"
)

type Manager struct {
	db      *gorm.DB
	scripts []Script
}

func NewManager(db *gorm.DB, scripts []Script) *Manager {
	return &Manager{
		db:      db,
		scripts: scripts,
	}
}

func (m *Manager) EnsureMigrationTable() error {
	return m.db.AutoMigrate(&Migration{})
}

func (m *Manager) isApplied(version string) (bool, error) {
	var count int64
	err := m.db.Model(&Migration{}).Where("version = ?", version).Count(&count).Error
	return count > 0, err
}

func (m *Manager) recordMigration(version, name string) error {
	migration := &Migration{
		Version: version,
		Name:    name,
	}
	return m.db.Create(migration).Error
}

func (m *Manager) removeMigration(version string) error {
	return m.db.Where("version = ?", version).Delete(&Migration{}).Error
}

func (m *Manager) Up() error {
	if err := m.EnsureMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied := 0
	for _, script := range m.scripts {
		isApplied, err := m.isApplied(script.Version())
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if isApplied {
			logger.Log.Infof("Migration %s (%s) already applied, skipping", script.Version(), script.Name())
			continue
		}

		logger.Log.Infof("Applying migration %s: %s", script.Version(), script.Name())

		if err := script.Up(m.db); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", script.Version(), err)
		}

		if err := m.recordMigration(script.Version(), script.Name()); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", script.Version(), err)
		}

		logger.Log.Infof("Successfully applied migration %s", script.Version())
		applied++
	}

	if applied == 0 {
		logger.Log.Info("No new migrations to apply")
	} else {
		logger.Log.Infof("Applied %d migration(s)", applied)
	}

	return nil
}

func (m *Manager) Down(steps int) error {
	if err := m.EnsureMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	if steps <= 0 {
		return errors.New("steps must be greater than 0")
	}

	var appliedMigrations []Migration
	err := m.db.Order("applied_at DESC").Limit(steps).Find(&appliedMigrations).Error
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(appliedMigrations) == 0 {
		logger.Log.Info("No migrations to rollback")
		return nil
	}

	rolled := 0
	for _, applied := range appliedMigrations {
		var script Script
		for _, s := range m.scripts {
			if s.Version() == applied.Version {
				script = s
				break
			}
		}

		if script == nil {
			logger.Log.Warnf("Migration script %s not found, removing from database", applied.Version)
			if err := m.removeMigration(applied.Version); err != nil {
				return fmt.Errorf("failed to remove migration record: %w", err)
			}
			continue
		}

		logger.Log.Infof("Rolling back migration %s: %s", script.Version(), script.Name())

		if err := script.Down(m.db); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", script.Version(), err)
		}

		if err := m.removeMigration(script.Version()); err != nil {
			return fmt.Errorf("failed to remove migration record: %w", err)
		}

		logger.Log.Infof("Successfully rolled back migration %s", script.Version())
		rolled++
	}

	logger.Log.Infof("Rolled back %d migration(s)", rolled)
	return nil
}

func (m *Manager) Fresh() error {
	logger.Log.Info("Running fresh migration (dropping all tables)...")

	if err := m.EnsureMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	var appliedMigrations []Migration
	if err := m.db.Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for i := len(appliedMigrations) - 1; i >= 0; i-- {
		applied := appliedMigrations[i]
		var script Script
		for _, s := range m.scripts {
			if s.Version() == applied.Version {
				script = s
				break
			}
		}

		if script != nil {
			logger.Log.Infof("Dropping tables for migration %s", script.Version())
			if err := script.Down(m.db); err != nil {
				logger.Log.Warnf("Failed to rollback migration %s: %v", script.Version(), err)
			}
		}
	}

	if err := m.db.Migrator().DropTable(&Migration{}); err != nil {
		logger.Log.Warnf("Failed to drop migrations table: %v", err)
	}

	logger.Log.Info("All tables dropped, running migrations...")
	return m.Up()
}

func (m *Manager) Status() error {
	if err := m.EnsureMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	var appliedMigrations []Migration
	if err := m.db.Order("applied_at ASC").Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]Migration)
	for _, m := range appliedMigrations {
		appliedMap[m.Version] = m
	}

	logger.Log.Info("Migration Status:")
	logger.Log.Info("=================")

	for _, script := range m.scripts {
		if applied, ok := appliedMap[script.Version()]; ok {
			logger.Log.Infof("[X] %s - %s (applied at: %s)", script.Version(), script.Name(), applied.AppliedAt.Format("2006-01-02 15:04:05"))
		} else {
			logger.Log.Infof("[ ] %s - %s (pending)", script.Version(), script.Name())
		}
	}

	return nil
}
