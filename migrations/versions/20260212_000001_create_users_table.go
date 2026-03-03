package versions

import (
	"gorm.io/gorm"
)

type Migration20260212_000001 struct{}

func (m Migration20260212_000001) Version() string {
	return "20260212_000001"
}

func (m Migration20260212_000001) Name() string {
	return "create_users_table"
}

func (m Migration20260212_000001) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			is_verified INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			last_login_at DATETIME
		)
	`).Error
}

func (m Migration20260212_000001) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS users").Error
}
