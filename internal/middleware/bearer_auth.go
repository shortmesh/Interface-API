package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"

	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type BearerAuthMiddleware struct {
	validator *TokenValidator
	config    *TokenConfig
}

func NewBearerAuth(db database.Service) *BearerAuthMiddleware {
	return &BearerAuthMiddleware{
		validator: NewTokenValidator(db.DB()),
		config:    NewTokenConfig(),
	}
}

func (m *BearerAuthMiddleware) Authenticate(methods ...AuthMethod) echo.MiddlewareFunc {
	if len(methods) == 0 {
		methods = []AuthMethod{AuthMethodSession}
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
			var matrixIdentity *models.MatrixIdentity
			var user *models.User
			var err error
			var isMatrixToken bool

			if strings.HasPrefix(token, m.config.SessionPrefix) {
				if !isMethodAllowed(methods, AuthMethodSession) {
					logger.Error("Session authentication not allowed for this endpoint")
					return echo.NewHTTPError(http.StatusUnauthorized, "session authentication not allowed")
				}
				session, err = m.validator.ValidateSessionToken(strings.TrimPrefix(token, m.config.SessionPrefix))
				if err == nil {
					user = &session.User
				}
			} else if strings.HasPrefix(token, m.config.MatrixPrefix) {
				if !isMethodAllowed(methods, AuthMethodMatrixToken) {
					logger.Error("Matrix token authentication not allowed for this endpoint")
					return echo.NewHTTPError(http.StatusUnauthorized, "matrix token authentication not allowed")
				}
				isMatrixToken = true
				matrixIdentity, err = m.validator.ValidateMatrixToken(strings.TrimPrefix(token, m.config.MatrixPrefix))
			} else {
				logger.Error(fmt.Sprintf(
					"Invalid token format: %s. Expected '%s...' or '%s...'",
					token, m.config.SessionPrefix, m.config.MatrixPrefix,
				))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
			}

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if isMatrixToken {
						logger.Error("Invalid or expired matrix token")
						return echo.NewHTTPError(http.StatusForbidden, "invalid or expired token")
					}
					logger.Error("Invalid or expired token")
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
				}
				logger.Error(fmt.Sprintf("Failed to authenticate:\n%v\n\n%s", err, debug.Stack()))
				return echo.NewHTTPError(http.StatusInternalServerError, "authentication failed")
			}

			if user != nil {
				c.Set("user", user)
			}
			if session != nil {
				c.Set("session", session)
			}
			if matrixIdentity != nil {
				c.Set("matrix_identity", matrixIdentity)
			}

			return next(c)
		}
	}
}

func (m *BearerAuthMiddleware) AuthenticateWebSocket(methods ...AuthMethod) echo.MiddlewareFunc {
	if len(methods) == 0 {
		methods = []AuthMethod{AuthMethodSession}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.QueryParam("token")
			if token == "" {
				logger.Error("Missing token query parameter")
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token query parameter")
			}

			var session *models.Session
			var matrixIdentity *models.MatrixIdentity
			var user *models.User
			var err error
			var isMatrixToken bool

			if strings.HasPrefix(token, m.config.SessionPrefix) {
				if !isMethodAllowed(methods, AuthMethodSession) {
					logger.Error("Session authentication not allowed for this endpoint")
					return echo.NewHTTPError(http.StatusUnauthorized, "session authentication not allowed")
				}
				session, err = m.validator.ValidateSessionToken(strings.TrimPrefix(token, m.config.SessionPrefix))
				if err == nil {
					user = &session.User
				}
			} else if strings.HasPrefix(token, m.config.MatrixPrefix) {
				if !isMethodAllowed(methods, AuthMethodMatrixToken) {
					logger.Error("Matrix token authentication not allowed for this endpoint")
					return echo.NewHTTPError(http.StatusUnauthorized, "matrix token authentication not allowed")
				}
				isMatrixToken = true
				matrixIdentity, err = m.validator.ValidateMatrixToken(strings.TrimPrefix(token, m.config.MatrixPrefix))
			} else {
				logger.Error(fmt.Sprintf(
					"Invalid token format: %s. Expected '%s...' or '%s...'",
					token, m.config.SessionPrefix, m.config.MatrixPrefix,
				))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
			}

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if isMatrixToken {
						logger.Error("Invalid or expired matrix token")
						return echo.NewHTTPError(http.StatusForbidden, "invalid or expired token")
					}
					logger.Error("Invalid or expired token")
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
				}
				logger.Error(fmt.Sprintf("Failed to authenticate:\n%v\n\n%s", err, debug.Stack()))
				return echo.NewHTTPError(http.StatusInternalServerError, "authentication failed")
			}

			if user != nil {
				c.Set("user", user)
			}
			if session != nil {
				c.Set("session", session)
			}
			if matrixIdentity != nil {
				c.Set("matrix_identity", matrixIdentity)
			}

			return next(c)
		}
	}
}

func isMethodAllowed(allowedMethods []AuthMethod, method AuthMethod) bool {
	return slices.Contains(allowedMethods, method)
}
