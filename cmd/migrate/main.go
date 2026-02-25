package main

import (
	"flag"
	"fmt"
	"os"

	"interface-api/internal/database"
	"interface-api/migrations"
	"interface-api/pkg/logger"
	"interface-api/pkg/migrator"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env.default", ".env")

	var (
		action string
		steps  int
	)

	flag.StringVar(&action, "action", "up", "Migration action: up, down, fresh, status")
	flag.IntVar(&steps, "steps", 1, "Number of migrations to rollback (only for 'down' action)")
	flag.Parse()

	logger.Log.Infof("Migration action: %s", action)

	db := database.New()
	scripts := migrations.GetAllMigrations()
	manager := migrator.NewManager(db.DB(), scripts)

	var err error
	switch action {
	case "up":
		err = manager.Up()
	case "down":
		err = manager.Down(steps)
	case "fresh":
		err = manager.Fresh()
	case "status":
		err = manager.Status()
	default:
		logger.Log.Fatalf("Invalid action: %s. Use: up, down, fresh, or status", action)
	}

	if err != nil {
		logger.Log.Fatalf("Migration failed: %v", err)
	}

	logger.Log.Info("Migration completed successfully")
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: migrate [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -action string\n")
		fmt.Fprintf(os.Stderr, "        Migration action: up, down, fresh, status (default \"up\")\n")
		fmt.Fprintf(os.Stderr, "  -steps int\n")
		fmt.Fprintf(os.Stderr, "        Number of migrations to rollback (default 1, only for 'down' action)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  migrate -action=up                    # Run all pending migrations\n")
		fmt.Fprintf(os.Stderr, "  migrate -action=down -steps=1         # Rollback last migration\n")
		fmt.Fprintf(os.Stderr, "  migrate -action=down -steps=3         # Rollback last 3 migrations\n")
		fmt.Fprintf(os.Stderr, "  migrate -action=fresh                 # Drop all tables and recreate\n")
		fmt.Fprintf(os.Stderr, "  migrate -action=status                # Show migration status\n\n")
		fmt.Fprintf(os.Stderr, "Environment Variables:\n")
		fmt.Fprintf(os.Stderr, "  AUTO_MIGRATE=true                     # Auto-run migrations on app start\n")
		fmt.Fprintf(os.Stderr, "  AUTO_CREATE_TABLES=true               # Auto-create tables on first run\n")
	}
}
