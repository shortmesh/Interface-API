package database

import (
	"fmt"
	"os"

	"interface-api/internal/database/models"
	"interface-api/pkg/crypto"
	"interface-api/pkg/logger"

	"gorm.io/gorm"
)

func InitializeSuperAdminCredentials(db *gorm.DB) error {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		logger.Warn("CLIENT_ID or CLIENT_SECRET not set in environment - skipping credential initialization")
		return nil
	}

	secretHash, err := crypto.Hash(clientSecret)
	if err != nil {
		return fmt.Errorf("failed to hash client secret: %w", err)
	}

	_, err = models.UpsertCredential(
		db,
		clientID,
		secretHash,
		models.RoleSuperAdmin,
		"System super admin credentials from environment",
	)
	if err != nil {
		return fmt.Errorf("failed to upsert credential: %w", err)
	}

	logger.Info("Super admin credentials initialized")
	return nil
}
