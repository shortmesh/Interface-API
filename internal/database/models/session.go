package models

import (
	"os"
	"strconv"
	"time"

	"interface-api/pkg/crypto"

	"gorm.io/gorm"
)

type Session struct {
	ID         uint
	UserID     uint
	TokenHash  []byte
	IPAddress  string
	UserAgent  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastUsedAt *time.Time

	User User `gorm:"foreignKey:UserID"`
}

func (Session) TableName() string {
	return "sessions"
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
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: now,
	}

	durationHours := 720
	if env := os.Getenv("SESSION_DURATION_HOURS"); env != "" {
		if parsed, err := strconv.Atoi(env); err == nil && parsed > 0 {
			durationHours = parsed
		}
	}
	session.ExpiresAt = now.Add(time.Duration(durationHours) * time.Hour)

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
		"token_hash":   session.TokenHash,
		"ip_address":   session.IPAddress,
		"user_agent":   session.UserAgent,
		"expires_at":   session.ExpiresAt,
		"created_at":   now,
		"last_used_at": nil,
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
