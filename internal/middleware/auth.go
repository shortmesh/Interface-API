package middleware

import (
	"net/http"
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

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				logger.Log.Error("Missing authorization header")
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 {
				logger.Log.Errorf("Invalid authorization header format: %s", authHeader)
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			scheme := strings.ToLower(parts[0])
			token := parts[1]

			if scheme != "bearer" {
				logger.Log.Errorf("Unsupported authorization scheme: %s", scheme)
				return echo.NewHTTPError(http.StatusUnauthorized, "unsupported authorization scheme")
			}

			var session *models.Session
			var err error

			if strings.HasPrefix(token, "sk_") {
				if !isMethodAllowed(methods, AuthMethodSession) {
					logger.Log.Error("Session authentication not allowed for this endpoint")
					return echo.NewHTTPError(http.StatusUnauthorized, "session authentication not allowed")
				}
				session, err = a.authenticateSession(strings.TrimPrefix(token, "sk_"))
			} else {
				logger.Log.Errorf("Invalid token format: %s. Expected 'sk_...'", token)
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
			}

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Log.Error("Invalid or expired token")
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
				}
				logger.Log.Errorf("Failed to authenticate session:\n%v\n\n%s", err, debug.Stack())
				return echo.NewHTTPError(http.StatusInternalServerError, "authentication failed")
			}

			c.Set("user", &session.User)
			c.Set("session", session)

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

func isMethodAllowed(allowedMethods []AuthMethod, method AuthMethod) bool {
	return slices.Contains(allowedMethods, method)
}
