package versions

import (
	"interface-api/internal/database/models"

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
	return db.AutoMigrate(&models.Session{})
}

func (m Migration20260212_000002) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.Session{})
}
