package models

import (
	"time"

	"gorm.io/gorm"
)

type MatrixProfile struct {
	ID             uint
	UserID         uint
	MatrixUsername string
	MatrixDeviceID string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	User User `gorm:"foreignKey:UserID"`
}

func (MatrixProfile) TableName() string {
	return "matrix_profiles"
}

func FindMatrixProfileByUserID(db *gorm.DB, userID uint) (*MatrixProfile, error) {
	var profile MatrixProfile
	err := db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}
