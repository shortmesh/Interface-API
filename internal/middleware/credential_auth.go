package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/crypto"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type CredentialAuthMiddleware struct {
	db *gorm.DB
}

func NewCredentialAuth(db database.Service) *CredentialAuthMiddleware {
	return &CredentialAuthMiddleware{
		db: db.DB(),
	}
}

func (m *CredentialAuthMiddleware) Authenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "basic" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}

			credParts := strings.SplitN(string(decoded), ":", 2)
			if len(credParts) != 2 {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}

			clientID := credParts[0]
			clientSecret := credParts[1]

			credential, err := models.FindCredentialByClientID(m.db, clientID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Error("Invalid client credentials")
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
				}
				logger.Error(fmt.Sprintf("Failed to find credential: %v", err))
				return echo.ErrInternalServerError
			}

			secretHash, err := crypto.Hash(clientSecret)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to hash client secret: %v", err))
				return echo.ErrInternalServerError
			}

			if subtle.ConstantTimeCompare(secretHash, credential.ClientSecret) != 1 {
				logger.Error("Invalid client credentials")
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}

			c.Set("credential", credential)
			return next(c)
		}
	}
}

func (m *CredentialAuthMiddleware) RequireScope(scope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			credential, ok := c.Get("credential").(*models.Credential)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authentication")
			}

			if !credential.HasScope(scope) {
				logger.Warn(fmt.Sprintf("Insufficient permissions for scope: %s", scope))
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(c)
		}
	}
}
