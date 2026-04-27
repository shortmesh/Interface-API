package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type AdminSession struct {
	ExpiresAt    time.Time
	MatrixToken  string
	CredentialID *uint
}

var (
	adminSessions     = make(map[string]*AdminSession)
	adminSessionMutex sync.RWMutex
)

func init() {
	go cleanupExpiredAdminSessions()
}

func cleanupExpiredAdminSessions() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		adminSessionMutex.Lock()
		now := time.Now()
		for token, session := range adminSessions {
			if now.After(session.ExpiresAt) {
				delete(adminSessions, token)
			}
		}
		adminSessionMutex.Unlock()
	}
}

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func StoreSession(token string, expiration time.Time, credentialID *uint) {
	adminSessionMutex.Lock()
	defer adminSessionMutex.Unlock()
	adminSessions[token] = &AdminSession{
		ExpiresAt:    expiration,
		MatrixToken:  "",
		CredentialID: credentialID,
	}
}

func IsSessionValid(token string) bool {
	adminSessionMutex.RLock()
	session, exists := adminSessions[token]
	adminSessionMutex.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		adminSessionMutex.Lock()
		delete(adminSessions, token)
		adminSessionMutex.Unlock()
		return false
	}
	return true
}

func ClearSession(token string) {
	adminSessionMutex.Lock()
	defer adminSessionMutex.Unlock()
	delete(adminSessions, token)
}

func GetMatrixToken(sessionToken string) string {
	adminSessionMutex.RLock()
	defer adminSessionMutex.RUnlock()
	if session, exists := adminSessions[sessionToken]; exists {
		return session.MatrixToken
	}
	return ""
}

func SetMatrixToken(sessionToken, matrixToken string) error {
	adminSessionMutex.Lock()
	defer adminSessionMutex.Unlock()
	if session, exists := adminSessions[sessionToken]; exists {
		session.MatrixToken = matrixToken
		return nil
	}
	return fmt.Errorf("session not found")
}

type AdminAuth struct {
	db *gorm.DB
}

func NewAdminAuth(db database.Service) *AdminAuth {
	return &AdminAuth{db: db.DB()}
}

func (a *AdminAuth) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("shortmesh_admin_token")
			if err != nil || !IsSessionValid(cookie.Value) {
				logger.Error("Invalid or missing admin session cookie")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or missing session. Please log in."})
			}
			return next(c)
		}
	}
}

func (a *AdminAuth) InjectMatrixToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("shortmesh_admin_token")
			if err != nil {
				logger.Error("Missing admin session cookie")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or missing session. Please log in."})
			}

			token := GetMatrixToken(cookie.Value)
			if token == "" {
				logger.Error("Matrix token not set for session")
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Matrix token not set for this session. Please set it first."})
			}

			matrixIdentity, err := models.FindMatrixIdentityByToken(a.db, token)
			if err != nil {
				logger.Error("Invalid matrix token in session")
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Invalid matrix token"})
			}

			c.Set("matrix_identity", matrixIdentity)
			c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			return next(c)
		}
	}
}

func (a *AdminAuth) InjectCredential() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("shortmesh_admin_token")
			if err != nil {
				logger.Error("Missing admin session cookie")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or missing session. Please log in."})
			}

			adminSessionMutex.RLock()
			session, exists := adminSessions[cookie.Value]
			adminSessionMutex.RUnlock()

			if !exists {
				logger.Error("Admin session not found for token")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or missing session. Please log in."})
			}

			if session.CredentialID == nil {
				logger.Error("No credential associated with admin session")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or missing session. Please log in."})
			}

			var credential models.Credential
			if err := a.db.First(&credential, *session.CredentialID).Error; err != nil {
				logger.Error(fmt.Sprintf("Failed to load credential for session: %v", err))
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or missing session. Please log in."})
			}

			c.Set("credential", &credential)
			return next(c)
		}
	}
}
