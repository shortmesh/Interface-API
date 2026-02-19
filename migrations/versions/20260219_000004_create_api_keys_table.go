package versions

import (
	"interface-api/internal/database/models"

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
	return db.AutoMigrate(&models.APIKey{})
}

func (m Migration20260219_000004) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.APIKey{})
}
