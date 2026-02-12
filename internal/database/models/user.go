package models

import (
	"errors"
	"time"

	"interface-api/pkg/crypto"
	"interface-api/pkg/password"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	Email        []byte     `gorm:"type:blob;not null" json:"-"`
	EmailHash    []byte     `gorm:"type:binary(32);uniqueIndex;not null" json:"-"`
	PasswordHash string     `gorm:"column:password_hash;not null;size:255" json:"-"`
	IsVerified   bool       `gorm:"default:false" json:"is_verified"`
	CreatedAt    time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"not null" json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	if u.PasswordHash != "" {
		if err := password.ValidatePassword(u.PasswordHash); err != nil {
			return err
		}

		hashedPassword, err := crypto.HashPassword(u.PasswordHash)
		if err != nil {
			return err
		}
		u.PasswordHash = hashedPassword
	}

	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = time.Now().UTC()
	return nil
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
	if err := password.ValidatePassword(plainPassword); err != nil {
		return err
	}

	hashedPassword, err := crypto.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = hashedPassword
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *User) RecordLogin() {
	now := time.Now().UTC()
	u.LastLoginAt = &now
}

func (u *User) SetEmail(email string) error {
	encrypted, err := crypto.Encrypt(email)
	if err != nil {
		return err
	}
	u.Email = encrypted

	hash, err := crypto.Hash(email)
	if err != nil {
		return err
	}
	u.EmailHash = hash

	return nil
}

func (u *User) GetEmail() (string, error) {
	return crypto.Decrypt(u.Email)
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
