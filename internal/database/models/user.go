package models

import (
	"errors"
	"time"

	"interface-api/pkg/crypto"

	"gorm.io/gorm"
)

type User struct {
	ID           uint
	Email        string
	PasswordHash string
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}

func (User) TableName() string {
	return "users"
}

func (u *User) ComparePassword(password string) error {
	match, err := crypto.VerifyPassword(password, u.PasswordHash)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("password does not match")
	}
	return nil
}

func (u *User) SetPassword(plainPassword string) error {
	hashedPassword, err := crypto.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = hashedPassword
	return nil
}

func (u *User) RecordLogin(db *gorm.DB) error {
	now := time.Now().UTC()
	u.LastLoginAt = &now
	return db.Model(&User{}).Where("id = ?", u.ID).Updates(map[string]any{
		"last_login_at": now,
		"updated_at":    now,
	}).Error
}

func FindUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	err := db.Where("email = ?", email).First(&user).Error
	return &user, err
}
