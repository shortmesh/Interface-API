package models

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"interface-api/pkg/crypto"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrOTPNotFound        = errors.New("OTP not found")
	ErrOTPExpired         = errors.New("OTP has expired")
	ErrOTPInvalidCode     = errors.New("invalid OTP code")
	ErrOTPTooManyAttempts = errors.New("too many attempts")
	ErrOTPInvalidated     = errors.New("OTP was invalidated")
)

type OTP struct {
	ID            uint
	UserServiceID uint
	CodeHash      []byte
	Identifier    string
	Platform      string
	Sender        string
	ExpiresAt     time.Time
	AttemptCount  int
	CreatedAt     time.Time

	UserService UserService `gorm:"foreignKey:UserServiceID"`
}

func (OTP) TableName() string {
	return "otps"
}

func (o *OTP) SetCode(code string) error {
	hash, err := crypto.Hash(code)
	if err != nil {
		return err
	}
	o.CodeHash = hash
	return nil
}

func (o *OTP) CompareCode(code string) error {
	hash, err := crypto.Hash(code)
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare(o.CodeHash, hash) != 1 {
		return errors.New("code does not match")
	}
	return nil
}

func GenerateOTP(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", err
	}

	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}

func CreateOTP(db *gorm.DB, userServiceID uint, identifier, platform, sender string) (string, time.Time, error) {
	otpLength := 6
	if lengthStr := os.Getenv("OTP_LENGTH"); lengthStr != "" {
		if length, err := strconv.Atoi(lengthStr); err == nil {
			if length < 6 {
				otpLength = 6
			} else if length > 12 {
				otpLength = 12
			} else {
				otpLength = length
			}
		}
	}

	expiryMinutes := 10
	if expiryStr := os.Getenv("OTP_EXPIRY_MINUTES"); expiryStr != "" {
		if expiry, err := strconv.Atoi(expiryStr); err == nil && expiry > 0 {
			expiryMinutes = expiry
		}
	}

	code, err := GenerateOTP(otpLength)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().UTC().Add(time.Duration(expiryMinutes) * time.Minute)

	otp := &OTP{
		UserServiceID: userServiceID,
		Identifier:    identifier,
		Platform:      platform,
		Sender:        sender,
		ExpiresAt:     expiresAt,
		AttemptCount:  0,
		CreatedAt:     time.Now().UTC(),
	}

	if err := otp.SetCode(code); err != nil {
		return "", time.Time{}, err
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_service_id"}, {Name: "identifier"}, {Name: "platform"}, {Name: "sender"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"code_hash", "expires_at", "attempt_count", "created_at"}),
	}).Create(otp).Error; err != nil {
		return "", time.Time{}, err
	}

	return code, expiresAt, nil
}

func VerifyOTP(db *gorm.DB, userServiceID uint, identifier, platform, sender, code string) error {
	maxAttempts := 3
	if maxAttemptsStr := os.Getenv("OTP_MAX_ATTEMPTS"); maxAttemptsStr != "" {
		if max, err := strconv.Atoi(maxAttemptsStr); err == nil && max > 0 {
			maxAttempts = max
		}
	}

	var otp OTP
	err := db.Where("user_service_id = ? AND identifier = ? AND platform = ? AND sender = ?",
		userServiceID, identifier, platform, sender).First(&otp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOTPNotFound
		}
		return err
	}

	if otp.ExpiresAt.Before(time.Now().UTC()) {
		db.Delete(&otp)
		return ErrOTPExpired
	}

	if err := db.Model(&otp).Update("attempt_count", gorm.Expr("attempt_count + 1")).Error; err != nil {
		return err
	}

	if err := db.First(&otp, otp.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOTPInvalidated
		}
		return err
	}

	if otp.AttemptCount > maxAttempts {
		db.Delete(&otp)
		return ErrOTPTooManyAttempts
	}

	if err := otp.CompareCode(code); err != nil {
		return ErrOTPInvalidCode
	}

	db.Delete(&otp)
	return nil
}
