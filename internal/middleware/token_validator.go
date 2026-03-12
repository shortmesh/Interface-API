package middleware

import (
	"interface-api/internal/database/models"

	"gorm.io/gorm"
)

type TokenValidator struct {
	db *gorm.DB
}

func NewTokenValidator(db *gorm.DB) *TokenValidator {
	return &TokenValidator{db: db}
}

func (tv *TokenValidator) ValidateSessionToken(token string) (*models.Session, error) {
	session, err := models.FindSessionByToken(tv.db, token)
	if err != nil {
		return nil, err
	}

	if err := session.UpdateLastUsed(tv.db); err != nil {
		return nil, err
	}

	return session, nil
}

func (tv *TokenValidator) ValidateMatrixToken(token string) (*models.MatrixIdentity, error) {
	matrixIdentity, err := models.FindMatrixIdentityByToken(tv.db, token)
	if err != nil {
		return nil, err
	}

	if err := matrixIdentity.UpdateLastUsed(tv.db); err != nil {
		return nil, err
	}

	return matrixIdentity, nil
}
