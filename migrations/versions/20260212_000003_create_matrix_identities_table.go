package versions

import (
	"gorm.io/gorm"
)

type Migration20260212_000003 struct{}

func (m Migration20260212_000003) Version() string {
	return "20260212_000003"
}

func (m Migration20260212_000003) Name() string {
	return "create_matrix_identities_table"
}

func (m Migration20260212_000003) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS matrix_identities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			matrix_username TEXT NOT NULL,
			matrix_device_id TEXT NOT NULL,
			token_hash BLOB NOT NULL UNIQUE,
			is_admin INTEGER DEFAULT 0,
			expires_at DATETIME,
			last_used_at DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`).Error
}

func (m Migration20260212_000003) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS matrix_identities").Error
}
