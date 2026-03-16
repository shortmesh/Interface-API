package migrations

import (
	"interface-api/migrations/versions"
	"interface-api/pkg/migrator"
)

func GetAllMigrations() []migrator.Script {
	return []migrator.Script{
		versions.Migration20260212_000003{},
	}
}
