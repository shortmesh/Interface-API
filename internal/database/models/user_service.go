package models

import (
	"crypto/subtle"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserService struct {
	ID           uint
	UserID       uint
	ServiceID    uint
	ClientID     string
	ClientSecret string
	IsEnabled    bool
	ExpiresAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time

	User    User    `gorm:"foreignKey:UserID"`
	Service Service `gorm:"foreignKey:ServiceID"`
}

func (UserService) TableName() string {
	return "user_services"
}

func (us *UserService) VerifyClientSecret(clientSecret string) bool {
	return subtle.ConstantTimeCompare([]byte(us.ClientSecret), []byte(clientSecret)) == 1
}

func (us *UserService) IsExpired() bool {
	if us.ExpiresAt == nil {
		return false
	}
	return us.ExpiresAt.Before(time.Now().UTC())
}

func FindUserServiceByUserAndService(db *gorm.DB, userID, serviceID uint) (*UserService, error) {
	var userService UserService
	err := db.Where("user_id = ? AND service_id = ?", userID, serviceID).
		Preload("User").
		Preload("Service").
		First(&userService).Error
	return &userService, err
}

func FindUserServicesByUser(db *gorm.DB, userID uint) ([]UserService, error) {
	var userServices []UserService
	err := db.Where("user_id = ?", userID).
		Preload("Service").
		Find(&userServices).Error
	return userServices, err
}

func CreateOrUpdateUserService(db *gorm.DB, userID, serviceID uint, clientID, clientSecret string, expiresAt *time.Time) (*UserService, error) {
	now := time.Now().UTC()

	userService := UserService{
		UserID:       userID,
		ServiceID:    serviceID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		IsEnabled:    true,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"}, {Name: "service_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"client_id", "client_secret", "is_enabled", "expires_at", "updated_at"}),
	}).Create(&userService).Error
	if err != nil {
		return nil, err
	}

	return &userService, nil
}
