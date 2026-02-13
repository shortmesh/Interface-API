package versions

import (
	"interface-api/internal/database/models"

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
	return db.AutoMigrate(&models.MatrixProfile{})
}

func (m Migration20260212_000003) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.MatrixProfile{})
}
