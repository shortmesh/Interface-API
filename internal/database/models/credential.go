package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CredentialRole string

const (
	RoleSuperAdmin CredentialRole = "super_admin"
	RoleUser       CredentialRole = "user"
)

type Scopes []string

func (s Scopes) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *Scopes) Scan(value any) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*s = []string{}
			return nil
		}
		bytes = []byte(str)
	}

	return json.Unmarshal(bytes, s)
}

var roleScopes = map[CredentialRole]Scopes{
	RoleSuperAdmin: {"*"},
	RoleUser: {
		"tokens:write:create",
		"devices:*",
		"webhooks:*",
	},
}

func GetDefaultScopesForRole(role CredentialRole) Scopes {
	if scopes, ok := roleScopes[role]; ok {
		result := make(Scopes, len(scopes))
		copy(result, scopes)
		return result
	}
	return Scopes{}
}

func ValidateScopesForRole(role CredentialRole, scopes Scopes) error {
	if role == RoleSuperAdmin {
		return nil
	}

	allowedScopes := roleScopes[role]
	for _, scope := range scopes {
		if !contains(allowedScopes, scope) {
			return fmt.Errorf("scope '%s' not allowed for role '%s'", scope, role)
		}
	}
	return nil
}

func contains(scopes Scopes, scope string) bool {
	for _, s := range scopes {
		if s == "*" || s == scope {
			return true
		}
		if strings.HasSuffix(s, ":*") {
			prefix := strings.TrimSuffix(s, "*")
			if strings.HasPrefix(scope, prefix) {
				return true
			}
		}
	}
	return false
}

type Credential struct {
	ID           uint           `json:"id"`
	ClientID     string         `json:"client_id" gorm:"uniqueIndex;not null"`
	ClientSecret []byte         `json:"-" gorm:"not null"`
	Role         CredentialRole `json:"role" gorm:"not null"`
	Scopes       Scopes         `json:"scopes" gorm:"type:text"`
	Description  string         `json:"description"`
	Active       bool           `json:"active" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func (Credential) TableName() string {
	return "credentials"
}

func (c *Credential) HasScope(scope string) bool {
	if c.Role == RoleSuperAdmin {
		return true
	}

	for _, s := range c.Scopes {
		if s == scope || s == "*" {
			return true
		}
		if strings.HasSuffix(s, ":*") {
			prefix := strings.TrimSuffix(s, "*")
			if strings.HasPrefix(scope, prefix) {
				return true
			}
		}
	}
	return false
}

func FindCredentialByClientID(db *gorm.DB, clientID string) (*Credential, error) {
	var credential Credential
	err := db.Where("client_id = ?", clientID).First(&credential).Error
	return &credential, err
}

func UpsertCredential(db *gorm.DB, clientID string, secretHash []byte, role CredentialRole, description string) (*Credential, error) {
	now := time.Now().UTC()
	credential := &Credential{
		ClientID:     clientID,
		ClientSecret: secretHash,
		Role:         role,
		Scopes:       GetDefaultScopesForRole(role),
		Description:  description,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "client_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"client_secret", "role", "scopes", "description", "active", "updated_at"}),
	}).Create(credential).Error

	return credential, err
}

func DeactivateCredential(db *gorm.DB, clientID string) error {
	return db.Model(&Credential{}).
		Where("client_id = ?", clientID).
		Update("active", false).Error
}

func FindActiveCredentials(db *gorm.DB) ([]Credential, error) {
	var credentials []Credential
	err := db.Where("active = ?", true).Find(&credentials).Error
	return credentials, err
}

func FindAllCredentials(db *gorm.DB) ([]Credential, error) {
	var credentials []Credential
	err := db.Find(&credentials).Error
	return credentials, err
}
