package versions

import (
	"gorm.io/gorm"
)

type Migration20260212_000003 struct{}

func (m Migration20260212_000003) Version() string {
	return "20260212_000003"
}

func (m Migration20260212_000003) Name() string {
	return "create_matrix_profiles_table"
}

func (m Migration20260212_000003) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS matrix_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			matrix_username TEXT NOT NULL,
			matrix_device_id TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`).Error
}

func (m Migration20260212_000003) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS matrix_profiles").Error
}
