package versions

import (
	"gorm.io/gorm"
)

type Migration20260423_000001 struct{}

func (m Migration20260423_000001) Version() string {
	return "20260423_000001"
}

func (m Migration20260423_000001) Name() string {
	return "create_credentials_table"
}

func (m Migration20260423_000001) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id TEXT NOT NULL UNIQUE,
			client_secret BLOB NOT NULL,
			role TEXT NOT NULL,
			scopes TEXT NOT NULL DEFAULT '[]',
			description TEXT,
			active INTEGER DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
		CREATE UNIQUE INDEX idx_credentials_client_id ON credentials(client_id);
		CREATE INDEX idx_credentials_active ON credentials(active);
		CREATE INDEX idx_credentials_role ON credentials(role);
	`).Error
}

func (m Migration20260423_000001) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS credentials").Error
}
