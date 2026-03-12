package commands

import (
	"flag"
	"fmt"
	"os"
	"time"

	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
)

func ServiceCommand(db database.Service, args []string) {
	if len(args) == 0 {
		printServiceUsage()
		os.Exit(1)
	}

	action := args[0]
	fs := flag.NewFlagSet("service", flag.ExitOnError)

	switch action {
	case "create":
		name := fs.String("name", "", "Service name (required)")
		displayName := fs.String("display-name", "", "Display name (required)")
		description := fs.String("description", "", "Service description")
		fs.Parse(args[1:])

		if *name == "" || *displayName == "" {
			logger.Error("name and display-name are required")
			os.Exit(1)
		}

		service := models.Service{
			Name:        *name,
			DisplayName: *displayName,
			Description: *description,
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := db.DB().Create(&service).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to create service: %v", err))
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("Service created: %s (ID: %d)", service.Name, service.ID))

	case "list":
		fs.Parse(args[1:])

		var services []models.Service
		if err := db.DB().Find(&services).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to list services: %v", err))
			os.Exit(1)
		}

		fmt.Printf("%-5s %-20s %-30s %-10s\n", "ID", "Name", "Display Name", "Active")
		fmt.Println("--------------------------------------------------------------------------------")
		for _, s := range services {
			activeStr := "Yes"
			if !s.IsActive {
				activeStr = "No"
			}
			fmt.Printf("%-5d %-20s %-30s %-10s\n", s.ID, s.Name, s.DisplayName, activeStr)
		}

	case "activate":
		id := fs.Uint("id", 0, "Service ID (required)")
		fs.Parse(args[1:])

		if *id == 0 {
			logger.Error("id is required")
			os.Exit(1)
		}

		if err := db.DB().Model(&models.Service{}).Where("id = ?", *id).Update("is_active", true).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to activate service: %v", err))
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("Service %d activated", *id))

	case "deactivate":
		id := fs.Uint("id", 0, "Service ID (required)")
		fs.Parse(args[1:])

		if *id == 0 {
			logger.Error("id is required")
			os.Exit(1)
		}

		if err := db.DB().Model(&models.Service{}).Where("id = ?", *id).Update("is_active", false).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to deactivate service: %v", err))
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("Service %d deactivated", *id))

	case "delete":
		id := fs.Uint("id", 0, "Service ID (required)")
		fs.Parse(args[1:])

		if *id == 0 {
			logger.Error("id is required")
			os.Exit(1)
		}

		if err := db.DB().Delete(&models.Service{}, *id).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to delete service: %v", err))
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("Service %d deleted", *id))

	default:
		logger.Error(fmt.Sprintf("Unknown service action: %s", action))
		printServiceUsage()
		os.Exit(1)
	}
}

func printServiceUsage() {
	fmt.Fprintf(os.Stderr, "Usage: admin service <action> [options]\n\n")
	fmt.Fprintf(os.Stderr, "Actions:\n")
	fmt.Fprintf(os.Stderr, "  create          Create a new service\n")
	fmt.Fprintf(os.Stderr, "  list            List all services\n")
	fmt.Fprintf(os.Stderr, "  activate        Activate a service\n")
	fmt.Fprintf(os.Stderr, "  deactivate      Deactivate a service\n")
	fmt.Fprintf(os.Stderr, "  delete          Delete a service\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  admin service create -name=authy -display-name=\"Authy OTP\" -description=\"2FA service\"\n")
	fmt.Fprintf(os.Stderr, "  admin service list\n")
	fmt.Fprintf(os.Stderr, "  admin service activate -id=1\n")
	fmt.Fprintf(os.Stderr, "  admin service deactivate -id=1\n")
	fmt.Fprintf(os.Stderr, "  admin service delete -id=1\n")
}
