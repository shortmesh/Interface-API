package apikeys

import (
	"net/http"
	"time"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// List godoc
//
//	@Summary		List API keys
//	@Description	Get all API keys for the authenticated user
//	@Tags			apikeys
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	ListAPIKeysResponse	"List of API keys"
//	@Failure		401	{object}	ErrorResponse		"Unauthorized"
//	@Failure		500	{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/api-keys [get]
func (h *APIKeyHandler) List(c echo.Context) error {
	user := c.Get("user").(*models.User)

	apiKeys, err := models.ListAPIKeys(h.db.DB(), user.ID)
	if err != nil {
		logger.Log.Errorf("Failed to list API keys: %v", err)
		return echo.ErrInternalServerError
	}

	response := ListAPIKeysResponse{
		APIKeys: make([]APIKeyInfo, len(apiKeys)),
	}

	for i, key := range apiKeys {
		var expiresAtStr *string
		if !key.ExpiresAt.IsZero() {
			str := key.ExpiresAt.Format(time.RFC3339)
			expiresAtStr = &str
		}

		var lastUsedAtStr *string
		if key.LastUsedAt != nil {
			str := key.LastUsedAt.Format(time.RFC3339)
			lastUsedAtStr = &str
		}

		response.APIKeys[i] = APIKeyInfo{
			KeyID:      key.KeyID,
			Name:       key.Name,
			ExpiresAt:  expiresAtStr,
			CreatedAt:  key.CreatedAt.Format(time.RFC3339),
			LastUsedAt: lastUsedAtStr,
		}
	}

	return c.JSON(http.StatusOK, response)
}
