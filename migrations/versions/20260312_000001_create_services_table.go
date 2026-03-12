package versions

import (
	"gorm.io/gorm"
)

type Migration20260312_000001 struct{}

func (m Migration20260312_000001) Version() string {
	return "20260312_000001"
}

func (m Migration20260312_000001) Name() string {
	return "create_services_table"
}

func (m Migration20260312_000001) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS services (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			description TEXT,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_services_active ON services(is_active);
	`).Error
}

func (m Migration20260312_000001) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS services").Error
}
