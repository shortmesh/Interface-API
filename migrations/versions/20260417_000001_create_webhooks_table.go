package versions

import (
	"gorm.io/gorm"
)

type Migration20260417_000001 struct{}

func (m Migration20260417_000001) Version() string {
	return "20260417_000001"
}

func (m Migration20260417_000001) Name() string {
	return "create_webhooks_table"
}

func (m Migration20260417_000001) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS webhooks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			matrix_identity_id INTEGER NOT NULL,
			url TEXT NOT NULL,
			active INTEGER DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (matrix_identity_id) REFERENCES matrix_identities(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_webhooks_matrix_identity_id ON webhooks(matrix_identity_id);
		CREATE UNIQUE INDEX idx_webhooks_identity_url ON webhooks(matrix_identity_id, url);
	`).Error
}

func (m Migration20260417_000001) Down(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS webhooks").Error
}
