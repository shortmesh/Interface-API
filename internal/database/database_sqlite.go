//go:build !sqlcipher

package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openDatabase(sqlitePath string, dbEncryptionKey string, gormConfig *gorm.Config) (*gorm.DB, error) {
	if dbEncryptionKey != "" {
		return nil, fmt.Errorf("DB_ENCRYPTION_KEY provided but encryption is disabled. Rebuild with sqlcipher tag or unset DISABLE_DB_ENCRYPTION")
	}
	return gorm.Open(sqlite.Open(sqlitePath), gormConfig)
}
