package versions

import (
	"gorm.io/gorm"
)

type Migration20260312_000003 struct{}

func (m Migration20260312_000003) Version() string {
	return "20260312_000003"
}

func (m Migration20260312_000003) Name() string {
	return "create_otps_table"
}

func (m Migration20260312_000003) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS otps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_service_id INTEGER NOT NULL,
			code_hash BLOB NOT NULL,
			identifier TEXT NOT NULL,
			platform TEXT NOT NULL,
			sender TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			attempt_count INTEGER DEFAULT 0 NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_service_id) REFERENCES user_services(id) ON DELETE CASCADE,
			UNIQUE(user_service_id, identifier, platform, sender)
		);
		CREATE INDEX IF NOT EXISTS idx_otps_expires_at ON otps(expires_at);
	`).Error
}

func (m Migration20260312_000003) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS otps").Error
}
