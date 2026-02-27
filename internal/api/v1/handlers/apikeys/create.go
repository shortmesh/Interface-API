package apikeys

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// Create godoc
//
//	@Summary		Create a new API key
//	@Description	Create a new API key for the authenticated user
//	@Tags			apikeys
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateAPIKeyRequest	true	"API key creation request"
//	@Success		201		{object}	APIKeyResponse		"API key created successfully"
//	@Failure		400		{object}	ErrorResponse		"Invalid request body or validation error"
//	@Failure		401		{object}	ErrorResponse		"Unauthorized"
//	@Failure		500		{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/api-keys [post]
func (h *APIKeyHandler) Create(c echo.Context) error {
	user := c.Get("user").(*models.User)

	var req CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("API key creation failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.Name) == "" {
		logger.Info("API key creation failed: missing name")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: name",
		})
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			logger.Info(fmt.Sprintf("API key creation failed: invalid expires_at format - %v", err))
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid expires_at format. Must be RFC3339 (e.g., 2026-12-31T23:59:59Z)",
			})
		}
		if parsedTime.Before(time.Now().UTC()) {
			logger.Info("API key creation failed: expires_at is in the past")
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "expires_at must be in the future",
			})
		}
		expiresAt = &parsedTime
	}

	token, apiKey, err := models.CreateAPIKey(h.db.DB(), user.ID, req.Name, expiresAt)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create API key: %v", err))
		return echo.ErrInternalServerError
	}

	var expiresAtStr *string
	if !apiKey.ExpiresAt.IsZero() {
		str := apiKey.ExpiresAt.Format(time.RFC3339)
		expiresAtStr = &str
	}

	logger.Info("API key created successfully")
	return c.JSON(http.StatusCreated, APIKeyResponse{
		Message: "API key created successfully",
		Key:     token,
		APIKey: &APIKeyInfo{
			KeyID:     apiKey.KeyID,
			Name:      apiKey.Name,
			ExpiresAt: expiresAtStr,
			CreatedAt: apiKey.CreatedAt.Format(time.RFC3339),
		},
	})
}
