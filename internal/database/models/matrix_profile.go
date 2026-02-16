package models

import (
	"time"

	"interface-api/pkg/crypto"

	"gorm.io/gorm"
)

type MatrixProfile struct {
	ID                       uint      `gorm:"primaryKey"`
	UserID                   uint      `gorm:"not null;uniqueIndex"`
	MatrixUsernameCiphertext []byte    `gorm:"type:blob;not null"`
	MatrixDeviceIDCiphertext []byte    `gorm:"type:blob;not null"`
	CreatedAt                time.Time `gorm:"not null"`
	UpdatedAt                time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (MatrixProfile) TableName() string {
	return "matrix_profiles"
}

func (m *MatrixProfile) SetMatrixUsername(username string) error {
	encrypted, err := crypto.Encrypt(username)
	if err != nil {
		return err
	}
	m.MatrixUsernameCiphertext = encrypted
	return nil
}

func (m *MatrixProfile) GetMatrixUsername() (string, error) {
	return crypto.Decrypt(m.MatrixUsernameCiphertext)
}

func (m *MatrixProfile) SetMatrixDeviceID(deviceID string) error {
	encrypted, err := crypto.Encrypt(deviceID)
	if err != nil {
		return err
	}
	m.MatrixDeviceIDCiphertext = encrypted
	return nil
}

func (m *MatrixProfile) GetMatrixDeviceID() (string, error) {
	return crypto.Decrypt(m.MatrixDeviceIDCiphertext)
}

func FindMatrixProfileByUserID(db *gorm.DB, userID uint) (*MatrixProfile, error) {
	var profile MatrixProfile
	err := db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}
