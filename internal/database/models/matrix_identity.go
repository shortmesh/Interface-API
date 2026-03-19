package models

import (
	"os"
	"strings"
	"time"

	"interface-api/pkg/crypto"

	"gorm.io/gorm"
)

type MatrixIdentity struct {
	ID             uint       `json:"id"`
	MatrixUsername string     `json:"matrix_username"`
	MatrixDeviceID string     `json:"matrix_device_id"`
	TokenHash      []byte     `json:"token_hash"`
	IsAdmin        bool       `json:"is_admin"`
	ExpiresAt      *time.Time `json:"expires_at"`
	LastUsedAt     *time.Time `json:"last_used_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (MatrixIdentity) TableName() string {
	return "matrix_identities"
}

func (m *MatrixIdentity) UpdateLastUsed(db *gorm.DB) error {
	now := time.Now().UTC()
	m.LastUsedAt = &now
	return db.Model(&MatrixIdentity{}).Where("id = ?", m.ID).Update("last_used_at", now).Error
}

func FindMatrixIdentityByToken(db *gorm.DB, token string) (*MatrixIdentity, error) {
	tokenPrefix := os.Getenv("MATRIX_TOKEN_PREFIX")
	if tokenPrefix == "" {
		tokenPrefix = "mt_"
	}

	if after, ok := strings.CutPrefix(token, tokenPrefix); ok {
		token = after
	}

	hash, err := crypto.Hash(token)
	if err != nil {
		return nil, err
	}

	var identity MatrixIdentity
	err = db.Where("token_hash = ? AND (expires_at IS NULL OR expires_at > ?)", hash, time.Now().UTC()).
		First(&identity).Error
	return &identity, err
}

func CreateMatrixIdentity(db *gorm.DB, matrixUsername, matrixDeviceID string, isAdmin bool, expiresAt *time.Time) (string, *MatrixIdentity, error) {
	tokenPrefix := os.Getenv("MATRIX_TOKEN_PREFIX")
	if tokenPrefix == "" {
		tokenPrefix = "mt_"
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

	identity := &MatrixIdentity{
		MatrixUsername: matrixUsername,
		MatrixDeviceID: matrixDeviceID,
		TokenHash:      hash,
		IsAdmin:        isAdmin,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if expiresAt != nil {
		identity.ExpiresAt = expiresAt
	}

	if err := db.Create(identity).Error; err != nil {
		return "", nil, err
	}

	return tokenPrefix + token, identity, nil
}

func FindAdminMatrixIdentity(db *gorm.DB) (*MatrixIdentity, error) {
	var identity MatrixIdentity
	err := db.Where("is_admin = ?", true).First(&identity).Error
	return &identity, err
}
