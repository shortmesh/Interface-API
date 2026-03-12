package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type BasicAuthMiddleware struct {
	db          database.Service
	validator   *TokenValidator
	config      *TokenConfig
	serviceName string
}

func NewBasicAuth(db database.Service, serviceName string) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{
		db:          db,
		validator:   NewTokenValidator(db.DB()),
		config:      NewTokenConfig(),
		serviceName: serviceName,
	}
}

func (m *BasicAuthMiddleware) Authenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userService, err := m.validateBasicAuth(c)
			if err != nil {
				return err
			}

			matrixIdentity, err := m.validateMatrixTokenHeader(c)
			if err != nil {
				return err
			}

			c.Set("user", &userService.User)
			c.Set("user_service", userService)
			c.Set("matrix_identity", matrixIdentity)

			return next(c)
		}
	}
}

func (m *BasicAuthMiddleware) validateBasicAuth(c echo.Context) (*models.UserService, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "basic" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	credParts := strings.SplitN(string(decoded), ":", 2)
	if len(credParts) != 2 {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	clientID := credParts[0]
	clientSecret := credParts[1]

	service, err := models.FindServiceByName(m.db.DB(), m.serviceName)
	if err != nil {
		logger.Error(fmt.Sprintf("Service not found: %v", err))
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	var userService models.UserService
	err = m.db.DB().Where("service_id = ? AND client_id = ?", service.ID, clientID).
		Preload("User").
		First(&userService).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}
		logger.Error(fmt.Sprintf("Database error: %v", err))
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "authentication failed")
	}

	if !userService.IsEnabled {
		return nil, echo.NewHTTPError(http.StatusForbidden, "service not enabled")
	}

	if userService.IsExpired() {
		return nil, echo.NewHTTPError(http.StatusForbidden, "service subscription expired")
	}

	if !userService.VerifyClientSecret(clientSecret) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	return &userService, nil
}

func (m *BasicAuthMiddleware) validateMatrixTokenHeader(c echo.Context) (*models.MatrixIdentity, error) {
	matrixTokenHeader := c.Request().Header.Get("X-Matrix-Token")
	if matrixTokenHeader == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "missing X-Matrix-Token header")
	}

	if !strings.HasPrefix(matrixTokenHeader, m.config.MatrixPrefix) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid matrix token format")
	}

	matrixToken := strings.TrimPrefix(matrixTokenHeader, m.config.MatrixPrefix)
	matrixIdentity, err := m.validator.ValidateMatrixToken(matrixToken)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusForbidden, "invalid or expired matrix token")
		}
		logger.Error(fmt.Sprintf("Failed to verify matrix token: %v", err))
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "authentication failed")
	}

	return matrixIdentity, nil
}
