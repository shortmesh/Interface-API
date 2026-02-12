package migrator

import (
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	ID        uint   `gorm:"primaryKey"`
	Version   string `gorm:"uniqueIndex;not null;size:255"`
	Name      string `gorm:"not null;size:255"`
	AppliedAt time.Time
}

func (Migration) TableName() string {
	return "migrations"
}

func (m *Migration) BeforeCreate(tx *gorm.DB) error {
	m.AppliedAt = time.Now().UTC()
	return nil
}

type Script interface {
	Version() string
	Name() string
	Up(*gorm.DB) error
	Down(*gorm.DB) error
}
