package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type BearerAuthMiddleware struct {
	db           *gorm.DB
	matrixPrefix string
}

func NewBearerAuth(db database.Service) *BearerAuthMiddleware {
	matrixPrefix := os.Getenv("MATRIX_TOKEN_PREFIX")
	if matrixPrefix == "" {
		matrixPrefix = "mt_"
	}
	return &BearerAuthMiddleware{
		db:           db.DB(),
		matrixPrefix: matrixPrefix,
	}
}

func (m *BearerAuthMiddleware) Authenticate() echo.MiddlewareFunc {
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

			if !strings.HasPrefix(token, m.matrixPrefix) {
				logger.Error(fmt.Sprintf("Invalid token format: %s. Expected '%s...'", token, m.matrixPrefix))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
			}

			matrixIdentity, err := m.validateMatrixToken(strings.TrimPrefix(token, m.matrixPrefix))
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Error("Invalid or expired matrix token")
					return echo.NewHTTPError(http.StatusForbidden, "invalid or expired token")
				}
				logger.Error(fmt.Sprintf("Failed to authenticate:\n%v\n\n%s", err, debug.Stack()))
				return echo.NewHTTPError(http.StatusInternalServerError, "authentication failed")
			}

			c.Set("matrix_identity", matrixIdentity)
			return next(c)
		}
	}
}

func (m *BearerAuthMiddleware) AuthenticateWebSocket() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.QueryParam("token")
			if token == "" {
				logger.Error("Missing token query parameter")
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token query parameter")
			}

			if !strings.HasPrefix(token, m.matrixPrefix) {
				logger.Error(fmt.Sprintf("Invalid token format: %s. Expected '%s...'", token, m.matrixPrefix))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
			}

			matrixIdentity, err := m.validateMatrixToken(strings.TrimPrefix(token, m.matrixPrefix))
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Error("Invalid or expired matrix token")
					return echo.NewHTTPError(http.StatusForbidden, "invalid or expired token")
				}
				logger.Error(fmt.Sprintf("Failed to authenticate:\n%v\n\n%s", err, debug.Stack()))
				return echo.NewHTTPError(http.StatusInternalServerError, "authentication failed")
			}

			c.Set("matrix_identity", matrixIdentity)
			return next(c)
		}
	}
}

func (m *BearerAuthMiddleware) validateMatrixToken(token string) (*models.MatrixIdentity, error) {
	return models.FindMatrixIdentityByToken(m.db, token)
}
