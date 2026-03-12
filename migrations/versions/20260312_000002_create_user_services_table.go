package versions

import (
	"gorm.io/gorm"
)

type Migration20260312_000002 struct{}

func (m Migration20260312_000002) Version() string {
	return "20260312_000002"
}

func (m Migration20260312_000002) Name() string {
	return "create_user_services_table"
}

func (m Migration20260312_000002) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS user_services (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			service_id INTEGER NOT NULL,
			client_id TEXT NOT NULL,
			client_secret TEXT NOT NULL,
			is_enabled INTEGER DEFAULT 1,
			expires_at DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
			UNIQUE(user_id, service_id)
		);
		CREATE INDEX IF NOT EXISTS idx_user_services_user ON user_services(user_id);
		CREATE INDEX IF NOT EXISTS idx_user_services_service ON user_services(service_id);
		CREATE INDEX IF NOT EXISTS idx_user_services_expires ON user_services(expires_at);
	`).Error
}

func (m Migration20260312_000002) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS user_services").Error
}
