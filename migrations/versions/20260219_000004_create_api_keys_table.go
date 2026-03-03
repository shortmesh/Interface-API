package versions

import (
	"gorm.io/gorm"
)

type Migration20260219_000004 struct{}

func (m Migration20260219_000004) Version() string {
	return "20260219_000004"
}

func (m Migration20260219_000004) Name() string {
	return "create_api_keys_table"
}

func (m Migration20260219_000004) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key_id TEXT NOT NULL UNIQUE,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			key_hash BLOB NOT NULL UNIQUE,
			expires_at DATETIME,
			created_at DATETIME NOT NULL,
			last_used_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);
		CREATE INDEX IF NOT EXISTS idx_api_keys_expires ON api_keys(expires_at);
	`).Error
}

func (m Migration20260219_000004) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS api_keys").Error
}
