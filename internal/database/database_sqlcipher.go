//go:build sqlcipher

package database

import (
	"fmt"

	sqlcipher "github.com/gdanko/gorm-sqlcipher"
	"gorm.io/gorm"
)

func openDatabase(sqlitePath string, dbEncryptionKey string, gormConfig *gorm.Config) (*gorm.DB, error) {
	if dbEncryptionKey == "" {
		return nil, fmt.Errorf("DB_ENCRYPTION_KEY is required when using SQLCipher encryption")
	}
	dsn := fmt.Sprintf(
		"%s?_pragma_key=x'%s'&_pragma_cipher_compatibility=3&_pragma_cipher_page_size=4096",
		sqlitePath,
		dbEncryptionKey,
	)
	return gorm.Open(sqlcipher.Open(dsn), gormConfig)
}
