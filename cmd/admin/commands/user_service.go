package commands

import (
	"flag"
	"fmt"
	"os"
	"time"

	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/crypto"
	"interface-api/pkg/logger"
)

func UserServiceCommand(db database.Service, args []string) {
	if len(args) == 0 {
		printUserServiceUsage()
		os.Exit(1)
	}

	action := args[0]
	fs := flag.NewFlagSet("user-service", flag.ExitOnError)

	switch action {
	case "grant":
		userID := fs.Uint("user-id", 0, "User ID (required)")
		serviceID := fs.Uint("service-id", 0, "Service ID (required)")
		clientID := fs.String("client-id", "", "Client ID (auto-generated if not provided)")
		clientSecret := fs.String("client-secret", "", "Client Secret (auto-generated if not provided)")
		expiresIn := fs.Int("expires-in-days", 0, "Expiration in days (0 for no expiration)")
		fs.Parse(args[1:])

		if *userID == 0 || *serviceID == 0 {
			logger.Error("user-id and service-id are required")
			os.Exit(1)
		}

		var expiresAt *time.Time
		if *expiresIn > 0 {
			expires := time.Now().UTC().Add(time.Duration(*expiresIn) * 24 * time.Hour)
			expiresAt = &expires
		}

		generatedClientID := *clientID
		generatedClientSecret := *clientSecret

		if generatedClientID == "" {
			token, err := crypto.GenerateSecureToken(16)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to generate client ID: %v", err))
				os.Exit(1)
			}
			generatedClientID = token
		}

		if generatedClientSecret == "" {
			token, err := crypto.GenerateSecureToken(32)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to generate client secret: %v", err))
				os.Exit(1)
			}
			generatedClientSecret = token
		}

		userService, err := models.CreateOrUpdateUserService(db.DB(), uint(*userID), uint(*serviceID), generatedClientID, generatedClientSecret, expiresAt)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to grant service access: %v", err))
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("Service access granted: UserService ID %d", userService.ID))
		logger.Info(fmt.Sprintf("Client ID: %s", userService.ClientID))
		logger.Info(fmt.Sprintf("Client Secret: %s", userService.ClientSecret))

	case "revoke":
		userID := fs.Uint("user-id", 0, "User ID (required)")
		serviceID := fs.Uint("service-id", 0, "Service ID (required)")
		fs.Parse(args[1:])

		if *userID == 0 || *serviceID == 0 {
			logger.Error("user-id and service-id are required")
			os.Exit(1)
		}

		if err := db.DB().Model(&models.UserService{}).
			Where("user_id = ? AND service_id = ?", *userID, *serviceID).
			Update("is_enabled", false).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to revoke service access: %v", err))
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("Service access revoked for user %d, service %d", *userID, *serviceID))

	case "enable":
		userID := fs.Uint("user-id", 0, "User ID (required)")
		serviceID := fs.Uint("service-id", 0, "Service ID (required)")
		fs.Parse(args[1:])

		if *userID == 0 || *serviceID == 0 {
			logger.Error("user-id and service-id are required")
			os.Exit(1)
		}

		if err := db.DB().Model(&models.UserService{}).
			Where("user_id = ? AND service_id = ?", *userID, *serviceID).
			Update("is_enabled", true).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to enable service: %v", err))
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("Service enabled for user %d, service %d", *userID, *serviceID))

	case "list":
		userID := fs.Uint("user-id", 0, "User ID (0 for all users)")
		fs.Parse(args[1:])

		var userServices []models.UserService
		query := db.DB().Preload("User").Preload("Service")

		if *userID > 0 {
			query = query.Where("user_id = ?", *userID)
		}

		if err := query.Find(&userServices).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to list user services: %v", err))
			os.Exit(1)
		}

		fmt.Printf("%-5s %-10s %-20s %-30s %-10s %-20s\n", "ID", "User ID", "Service", "Client ID", "Enabled", "Expires")
		fmt.Println("-------------------------------------------------------------------------------------------------------")
		for _, us := range userServices {
			enabledStr := "Yes"
			if !us.IsEnabled {
				enabledStr = "No"
			}

			expiresStr := "Never"
			if us.ExpiresAt != nil {
				expiresStr = us.ExpiresAt.Format("2006-01-02 15:04")
			}

			serviceName := ""
			if us.Service.Name != "" {
				serviceName = us.Service.Name
			}

			fmt.Printf("%-5d %-10d %-20s %-30s %-10s %-20s\n",
				us.ID, us.UserID, serviceName, us.ClientID, enabledStr, expiresStr)
		}

	case "delete":
		id := fs.Uint("id", 0, "UserService ID (required)")
		fs.Parse(args[1:])

		if *id == 0 {
			logger.Error("id is required")
			os.Exit(1)
		}

		if err := db.DB().Delete(&models.UserService{}, *id).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to delete user service: %v", err))
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("User service %d deleted", *id))

	default:
		logger.Error(fmt.Sprintf("Unknown user-service action: %s", action))
		printUserServiceUsage()
		os.Exit(1)
	}
}

func printUserServiceUsage() {
	fmt.Fprintf(os.Stderr, "Usage: admin user-service <action> [options]\n\n")
	fmt.Fprintf(os.Stderr, "Actions:\n")
	fmt.Fprintf(os.Stderr, "  grant           Grant service access to a user\n")
	fmt.Fprintf(os.Stderr, "  revoke          Revoke service access from a user\n")
	fmt.Fprintf(os.Stderr, "  enable          Enable service access for a user\n")
	fmt.Fprintf(os.Stderr, "  list            List user service subscriptions\n")
	fmt.Fprintf(os.Stderr, "  delete          Delete a user service subscription\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  admin user-service grant -user-id=1 -service-id=1  # Auto-generates credentials\n")
	fmt.Fprintf(os.Stderr, "  admin user-service grant -user-id=1 -service-id=1 -expires-in-days=365\n")
	fmt.Fprintf(os.Stderr, "  admin user-service grant -user-id=1 -service-id=1 -client-id=abc123 -client-secret=secret456\n")
	fmt.Fprintf(os.Stderr, "  admin user-service revoke -user-id=1 -service-id=1\n")
	fmt.Fprintf(os.Stderr, "  admin user-service enable -user-id=1 -service-id=1\n")
	fmt.Fprintf(os.Stderr, "  admin user-service list\n")
	fmt.Fprintf(os.Stderr, "  admin user-service list -user-id=1\n")
	fmt.Fprintf(os.Stderr, "  admin user-service delete -id=1\n")
}
