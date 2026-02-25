# Migrations

Database migration guide.

## Usage

```bash
make migrate-up          # Run pending migrations
make migrate-status      # Show status
make migrate-down        # Rollback last migration
```

Or use the binary:

```bash
go run cmd/migrate/main.go -action=up
go run cmd/migrate/main.go -action=down -steps=3
```

## Creating Migrations

1. Create file in `versions/`: `YYYYMMDD_XXXXXX_description.go`

```go
package versions

import (
  "interface-api/internal/database/models"
  "gorm.io/gorm"
)

type Migration20240212_000002 struct{}

func (m Migration20260212_000002) Version() string {
  return "20240212_000002"
}

func (m Migration20260212_000002) Name() string {
  return "add_profile_table"
}

func (m Migration20260212_000002) Up(db *gorm.DB) error {
  return db.AutoMigrate(&models.Profile{})
}

func (m Migration20260212_000002) Down(db *gorm.DB) error {
  return db.Migrator().DropTable(&models.Profile{})
}
```

2. Register in `migrations.go`:

```go
func GetAllMigrations() []migrator.Script {
  return []migrator.Script{
  versions.Migration20260212_000001{},
  versions.Migration20260212_000002{},
  }
}
```
