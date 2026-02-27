package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"slices"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (a *AuthMiddleware) Authenticate(methods ...AuthMethod) echo.MiddlewareFunc {
	if len(methods) == 0 {
		methods = []AuthMethod{AuthMethodSession}
	}

	sessionTokenPrefix := os.Getenv("SESSION_TOKEN_PREFIX")
	if sessionTokenPrefix == "" {
		sessionTokenPrefix = "sk_"
	}

	apiKeyPrefix := os.Getenv("API_KEY_PREFIX")
	if apiKeyPrefix == "" {
		apiKeyPrefix = "ak_"
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				logger.Error("Missing authorization header")
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 {
				logger.Error(fmt.Sprintf("Invalid authorization header format: %s", authHeader))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			scheme := strings.ToLower(parts[0])
			token := parts[1]

			if scheme != "bearer" {
				logger.Error(fmt.Sprintf("Unsupported authorization scheme: %s", scheme))
				return echo.NewHTTPError(http.StatusUnauthorized, "unsupported authorization scheme")
			}

			var session *models.Session
			var apiKey *models.APIKey
			var user *models.User
			var err error

			if strings.HasPrefix(token, sessionTokenPrefix) {
				if !isMethodAllowed(methods, AuthMethodSession) {
					logger.Error("Session authentication not allowed for this endpoint")
					return echo.NewHTTPError(http.StatusUnauthorized, "session authentication not allowed")
				}
				session, err = a.authenticateSession(strings.TrimPrefix(token, sessionTokenPrefix))
				if err == nil {
					user = &session.User
				}
			} else if strings.HasPrefix(token, apiKeyPrefix) {
				if !isMethodAllowed(methods, AuthMethodAPIKey) {
					logger.Error("API key authentication not allowed for this endpoint")
					return echo.NewHTTPError(http.StatusUnauthorized, "api key authentication not allowed")
				}
				apiKey, err = a.authenticateAPIKey(strings.TrimPrefix(token, apiKeyPrefix))
				if err == nil {
					user = &apiKey.User
				}
			} else {
				logger.Error(fmt.Sprintf(
					"Invalid token format: %s. Expected '%s...' or '%s...'",
					sessionTokenPrefix, apiKeyPrefix, token,
				))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
			}

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Error("Invalid or expired token")
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
				}
				logger.Error(fmt.Sprintf("Failed to authenticate:\n%v\n\n%s", err, debug.Stack()))
				return echo.NewHTTPError(http.StatusInternalServerError, "authentication failed")
			}

			c.Set("user", user)
			if session != nil {
				c.Set("session", session)
			}
			if apiKey != nil {
				c.Set("api_key", apiKey)
			}

			return next(c)
		}
	}
}

func (a *AuthMiddleware) authenticateSession(token string) (*models.Session, error) {
	session, err := models.FindSessionByToken(a.db.DB(), token)
	if err != nil {
		return nil, err
	}

	if err := session.UpdateLastUsed(a.db.DB()); err != nil {
		return nil, err
	}

	return session, nil
}

func (a *AuthMiddleware) authenticateAPIKey(token string) (*models.APIKey, error) {
	apiKey, err := models.FindAPIKeyByToken(a.db.DB(), token)
	if err != nil {
		return nil, err
	}

	if err := apiKey.UpdateLastUsed(a.db.DB()); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func isMethodAllowed(allowedMethods []AuthMethod, method AuthMethod) bool {
	return slices.Contains(allowedMethods, method)
}
