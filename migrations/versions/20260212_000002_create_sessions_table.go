package versions

import (
	"gorm.io/gorm"
)

type Migration20260212_000002 struct{}

func (m Migration20260212_000002) Version() string {
	return "20260212_000002"
}

func (m Migration20260212_000002) Name() string {
	return "create_sessions_table"
}

func (m Migration20260212_000002) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token_hash BLOB NOT NULL UNIQUE,
			ip_address TEXT,
			user_agent TEXT,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			last_used_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_user_expires ON sessions(user_id, expires_at);
		CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
	`).Error
}

func (m Migration20260212_000002) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS sessions").Error
}
