package migrations

import (
	"interface-api/migrations/versions"
	"interface-api/pkg/migrator"
)

func GetAllMigrations() []migrator.Script {
	return []migrator.Script{
		versions.Migration_20240212_000001{},
		versions.Migration_20240212_000002{},
	}
}
