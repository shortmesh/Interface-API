package versions

import (
	"time"

	"gorm.io/gorm"
)

type Migration20260312_000004 struct{}

func (m Migration20260312_000004) Version() string {
	return "20260312_000004"
}

func (m Migration20260312_000004) Name() string {
	return "insert_authy_service"
}

func (m Migration20260312_000004) Up(db *gorm.DB) error {
	now := time.Now().UTC()
	return db.Exec(`
		INSERT INTO services (name, display_name, description, is_active, created_at, updated_at)
		VALUES ('authy', 'Authy', 'Two-Factor Authentication Service', 1, ?, ?)
	`, now, now).Error
}

func (m Migration20260312_000004) Down(db *gorm.DB) error {
	return db.Exec("DELETE FROM services WHERE name = 'authy'").Error
}
