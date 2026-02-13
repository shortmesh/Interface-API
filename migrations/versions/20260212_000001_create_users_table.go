package versions

import (
	"interface-api/internal/database/models"

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
	return db.AutoMigrate(&models.User{})
}

func (m Migration20260212_000001) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.User{})
}
