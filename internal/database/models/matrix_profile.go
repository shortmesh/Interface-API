package models

import (
	"time"

	"interface-api/pkg/crypto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MatrixProfile struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:char(36);not null;uniqueIndex" json:"user_id"`
	MatrixUsername []byte    `gorm:"type:blob;not null" json:"-"`
	MatrixDeviceID []byte    `gorm:"type:blob;not null" json:"-"`
	CreatedAt      time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null" json:"updated_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (MatrixProfile) TableName() string {
	return "matrix_profiles"
}

func (m *MatrixProfile) BeforeCreate(tx *gorm.DB) error {
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MatrixProfile) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MatrixProfile) SetMatrixUsername(username string) error {
	encrypted, err := crypto.Encrypt(username)
	if err != nil {
		return err
	}
	m.MatrixUsername = encrypted
	return nil
}

func (m *MatrixProfile) GetMatrixUsername() (string, error) {
	return crypto.Decrypt(m.MatrixUsername)
}

func (m *MatrixProfile) SetMatrixDeviceID(deviceID string) error {
	encrypted, err := crypto.Encrypt(deviceID)
	if err != nil {
		return err
	}
	m.MatrixDeviceID = encrypted
	return nil
}

func (m *MatrixProfile) GetMatrixDeviceID() (string, error) {
	return crypto.Decrypt(m.MatrixDeviceID)
}

func FindMatrixProfileByUserID(db *gorm.DB, userID uuid.UUID) (*MatrixProfile, error) {
	var profile MatrixProfile
	err := db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}
