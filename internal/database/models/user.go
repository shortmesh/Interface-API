package models

import (
	"errors"
	"time"

	"interface-api/pkg/crypto"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
)

type User struct {
	ID              uint      `gorm:"primaryKey"`
	EmailCiphertext []byte    `gorm:"type:blob;not null"`
	EmailHash       []byte    `gorm:"type:binary(32);uniqueIndex;not null"`
	PasswordHash    string    `gorm:"not null;size:255"`
	IsVerified      bool      `gorm:"default:false"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
	LastLoginAt     *time.Time
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

func (u *User) SetEmail(email string) error {
	encrypted, err := crypto.Encrypt(email)
	if err != nil {
		return err
	}
	u.EmailCiphertext = encrypted

	hash, err := crypto.Hash(email)
	if err != nil {
		return err
	}
	u.EmailHash = hash

	return nil
}

func (u *User) GetEmail() (string, error) {
	return crypto.Decrypt(u.EmailCiphertext)
}

func FindUserByEmail(db *gorm.DB, email string) (*User, error) {
	hash, err := crypto.Hash(email)
	if err != nil {
		return nil, err
	}

	var user User
	err = db.Where("email_hash = ?", hash).First(&user).Error
	return &user, err
}
