package models

import (
	"os"
	"strconv"
	"time"

	"interface-api/pkg/crypto"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
)

type Session struct {
	ID         uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:char(36);not null;index:idx_user_expires" json:"user_id"`
	Token      string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	IPAddress  []byte     `gorm:"type:blob" json:"-"`
	UserAgent  []byte     `gorm:"type:blob" json:"-"`
	ExpiresAt  time.Time  `gorm:"not null;index:idx_user_expires" json:"-"`
	CreatedAt  time.Time  `gorm:"not null" json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Session) TableName() string {
	return "sessions"
}

func (s *Session) SetIPAddress(ip string) error {
	encrypted, err := crypto.Encrypt(ip)
	if err != nil {
		return err
	}
	s.IPAddress = encrypted
	return nil
}

func (s *Session) SetUserAgent(userAgent string) error {
	encrypted, err := crypto.Encrypt(userAgent)
	if err != nil {
		return err
	}
	s.UserAgent = encrypted
	return nil
}

func (s *Session) UpdateLastUsed(db *gorm.DB) error {
	now := time.Now().UTC()
	s.LastUsedAt = &now
	return db.Model(s).Update("last_used_at", now).Error
}

func FindSessionByToken(db *gorm.DB, token string, userID uuid.UUID) (*Session, error) {
	var session Session
	err := db.Where("token = ? AND user_id = ? AND expires_at > ?", token, userID, time.Now().UTC()).
		First(&session).Error
	return &session, err
}

func CreateOrUpdateSession(db *gorm.DB, userID uuid.UUID, ipAddress, userAgent string) (*Session, error) {
	maxSessions := getMaxSessionsPerUser()
	now := time.Now().UTC()

	session := &Session{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: now,
	}

	token, err := crypto.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}
	session.Token = token

	durationHours := 720
	if envDuration := os.Getenv("SESSION_DURATION_HOURS"); envDuration != "" {
		if parsed, err := strconv.Atoi(envDuration); err == nil && parsed > 0 {
			durationHours = parsed
		}
	}
	session.ExpiresAt = now.Add(time.Duration(durationHours) * time.Hour)

	if err := session.SetIPAddress(ipAddress); err != nil {
		return nil, err
	}
	if err := session.SetUserAgent(userAgent); err != nil {
		return nil, err
	}

	db.Where("user_id = ? AND expires_at <= ?", userID, now).Delete(&Session{})

	var count int64
	if err := db.Model(&Session{}).
		Where("user_id = ? AND expires_at > ?", userID, now).
		Count(&count).Error; err != nil {
		return nil, err
	}

	if int(count) < maxSessions {
		if err := db.Create(session).Error; err != nil {
			return nil, err
		}
		return session, nil
	}

	var oldest Session
	if err := db.Where("user_id = ? AND expires_at > ?", userID, now).
		Order("created_at ASC").
		First(&oldest).Error; err != nil {
		return nil, err
	}

	session.ID = oldest.ID

	if err := db.Model(&Session{}).Where("id = ?", oldest.ID).Updates(map[string]any{
		"token":        session.Token,
		"ip_address":   session.IPAddress,
		"user_agent":   session.UserAgent,
		"expires_at":   session.ExpiresAt,
		"created_at":   now,
		"last_used_at": nil,
	}).Error; err != nil {
		return nil, err
	}

	return session, nil
}

func getMaxSessionsPerUser() int {
	if val := os.Getenv("MAX_SESSIONS_PER_USER"); val != "" {
		if max, err := strconv.Atoi(val); err == nil && max > 0 {
			return max
		}
	}
	return 1
}
