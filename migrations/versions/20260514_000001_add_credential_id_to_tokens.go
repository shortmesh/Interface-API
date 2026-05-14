package versions

import (
	"gorm.io/gorm"
)

type Migration20260514_000001 struct{}

func (m Migration20260514_000001) Version() string {
	return "20260514_000001"
}

func (m Migration20260514_000001) Name() string {
	return "add_credential_id_to_matrix_identities"
}

func (m Migration20260514_000001) Up(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE matrix_identities ADD COLUMN credential_id INTEGER;
		CREATE INDEX idx_matrix_identities_credential_id ON matrix_identities(credential_id);
	`).Error
}

func (m Migration20260514_000001) Down(db *gorm.DB) error {
	return db.Exec(`
		DROP INDEX IF EXISTS idx_matrix_identities_credential_id;
		ALTER TABLE matrix_identities DROP COLUMN credential_id;
	`).Error
}
