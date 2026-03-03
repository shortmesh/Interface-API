package models

import (
	"encoding/hex"
	"os"
	"time"

	"interface-api/pkg/crypto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIKey struct {
	ID         uint
	KeyID      string
	UserID     uint
	Name       string
	KeyHash    []byte
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastUsedAt *time.Time

	User User `gorm:"foreignKey:UserID"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

func (a *APIKey) UpdateLastUsed(db *gorm.DB) error {
	now := time.Now().UTC()
	a.LastUsedAt = &now
	return db.Model(&APIKey{}).Where("id = ?", a.ID).Update("last_used_at", now).Error
}

func FindAPIKeyByToken(db *gorm.DB, token string) (*APIKey, error) {
	hash, err := crypto.Hash(token)
	if err != nil {
		return nil, err
	}

	var apiKey APIKey
	err = db.Where("key_hash = ? AND (expires_at IS NULL OR expires_at > ?)", hash, time.Now().UTC()).
		Preload("User").
		First(&apiKey).Error
	return &apiKey, err
}

func CreateAPIKey(db *gorm.DB, userID uint, name string, expiresAt *time.Time) (string, *APIKey, error) {
	apiKeyPrefix := os.Getenv("API_KEY_PREFIX")
	if apiKeyPrefix == "" {
		apiKeyPrefix = "ak_"
	}

	now := time.Now().UTC()

	token, err := crypto.GenerateSecureToken(32)
	if err != nil {
		return "", nil, err
	}

	hash, err := crypto.Hash(token)
	if err != nil {
		return "", nil, err
	}

	u := uuid.New()
	keyID := hex.EncodeToString(u[:])[:16]

	apiKey := &APIKey{
		KeyID:     keyID,
		UserID:    userID,
		Name:      name,
		KeyHash:   hash,
		CreatedAt: now,
	}

	if expiresAt != nil {
		apiKey.ExpiresAt = *expiresAt
	}

	if err := db.Create(apiKey).Error; err != nil {
		return "", nil, err
	}

	return apiKeyPrefix + token, apiKey, nil
}

func ListAPIKeys(db *gorm.DB, userID uint) ([]APIKey, error) {
	var apiKeys []APIKey
	err := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&apiKeys).Error
	return apiKeys, err
}

func DeleteAPIKey(db *gorm.DB, userID uint, keyID string) error {
	return db.Where("key_id = ? AND user_id = ?", keyID, userID).Delete(&APIKey{}).Error
}
