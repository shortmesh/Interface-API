package main

import (
	"fmt"
	"os"

	"interface-api/cmd/admin/commands"
	"interface-api/internal/database"
	_ "interface-api/pkg/config"
	"interface-api/pkg/logger"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	db := database.New()

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "service":
		commands.ServiceCommand(db, args)
	case "user-service":
		commands.UserServiceCommand(db, args)
	default:
		logger.Error(fmt.Sprintf("Unknown command: %s", command))
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: admin <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  service         Manage services\n")
	fmt.Fprintf(os.Stderr, "  user-service    Manage user service subscriptions\n\n")
	fmt.Fprintf(os.Stderr, "Run 'admin <command>' for more information on a command.\n")
}
