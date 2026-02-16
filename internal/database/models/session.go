package models

import (
	"os"
	"strconv"
	"time"

	"interface-api/pkg/crypto"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
)

type Session struct {
	ID                  uint      `gorm:"primaryKey"`
	UserID              uint      `gorm:"not null;index:idx_user_expires"`
	TokenHash           []byte    `gorm:"type:binary(32);not null;uniqueIndex"`
	IPAddressCiphertext []byte    `gorm:"type:blob"`
	UserAgentCiphertext []byte    `gorm:"type:blob"`
	ExpiresAt           time.Time `gorm:"not null;index:idx_user_expires;index"`
	CreatedAt           time.Time `gorm:"not null"`
	LastUsedAt          *time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (Session) TableName() string {
	return "sessions"
}

func (s *Session) SetIPAddress(ip string) error {
	encrypted, err := crypto.Encrypt(ip)
	if err != nil {
		return err
	}
	s.IPAddressCiphertext = encrypted
	return nil
}

func (s *Session) SetUserAgent(userAgent string) error {
	encrypted, err := crypto.Encrypt(userAgent)
	if err != nil {
		return err
	}
	s.UserAgentCiphertext = encrypted
	return nil
}

func (s *Session) UpdateLastUsed(db *gorm.DB) error {
	now := time.Now().UTC()
	s.LastUsedAt = &now
	return db.Model(&Session{}).Where("id = ?", s.ID).Update("last_used_at", now).Error
}

func FindSessionByToken(db *gorm.DB, token string) (*Session, error) {
	hash, err := crypto.Hash(token)
	if err != nil {
		return nil, err
	}

	var session Session
	err = db.Where("token_hash = ? AND expires_at > ?", hash, time.Now().UTC()).
		Preload("User").
		First(&session).Error
	return &session, err
}

func CreateOrUpdateSession(db *gorm.DB, userID uint, ipAddress, userAgent string) (string, error) {
	sessionTokenPrefix := os.Getenv("SESSION_TOKEN_PREFIX")
	if sessionTokenPrefix == "" {
		sessionTokenPrefix = "sk_"
	}

	now := time.Now().UTC()
	maxSessions := getMaxSessionsPerUser()

	token, err := crypto.GenerateSecureToken(32)
	if err != nil {
		return "", err
	}
	hash, err := crypto.Hash(token)
	if err != nil {
		return "", err
	}

	session := &Session{
		UserID:    userID,
		TokenHash: hash,
		CreatedAt: now,
	}

	durationHours := 720
	if env := os.Getenv("SESSION_DURATION_HOURS"); env != "" {
		if parsed, err := strconv.Atoi(env); err == nil && parsed > 0 {
			durationHours = parsed
		}
	}
	session.ExpiresAt = now.Add(time.Duration(durationHours) * time.Hour)

	if err := session.SetIPAddress(ipAddress); err != nil {
		return "", err
	}
	if err := session.SetUserAgent(userAgent); err != nil {
		return "", err
	}

	if err := db.Where("user_id = ? AND expires_at <= ?", userID, now).
		Delete(&Session{}).Error; err != nil {
		return "", err
	}

	var count int64
	if err := db.Model(&Session{}).
		Where("user_id = ? AND expires_at > ?", userID, now).
		Count(&count).Error; err != nil {
		return "", err
	}

	if int(count) < maxSessions {
		if err := db.Create(session).Error; err != nil {
			return "", err
		}
		return sessionTokenPrefix + token, nil
	}

	var oldest Session
	if err := db.Where("user_id = ? AND expires_at > ?", userID, now).
		Order("created_at ASC").
		First(&oldest).Error; err != nil {
		return "", err
	}

	if err := db.Model(&Session{}).Where("id = ?", oldest.ID).Updates(map[string]any{
		"token_hash":            session.TokenHash,
		"ip_address_ciphertext": session.IPAddressCiphertext,
		"user_agent_ciphertext": session.UserAgentCiphertext,
		"expires_at":            session.ExpiresAt,
		"created_at":            now,
		"last_used_at":          nil,
	}).Error; err != nil {
		return "", err
	}

	return sessionTokenPrefix + token, nil
}

func getMaxSessionsPerUser() int {
	if val := os.Getenv("MAX_SESSIONS_PER_USER"); val != "" {
		if max, err := strconv.Atoi(val); err == nil && max > 0 {
			return max
		}
	}
	return 1
}
