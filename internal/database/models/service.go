package models

import (
	"time"

	"gorm.io/gorm"
)

type Service struct {
	ID          uint
	Name        string
	DisplayName string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Service) TableName() string {
	return "services"
}

func FindServiceByName(db *gorm.DB, name string) (*Service, error) {
	var service Service
	err := db.Where("name = ? AND is_active = ?", name, true).First(&service).Error
	return &service, err
}

func FindAllActiveServices(db *gorm.DB) ([]Service, error) {
	var services []Service
	err := db.Where("is_active = ?", true).Find(&services).Error
	return services, err
}
