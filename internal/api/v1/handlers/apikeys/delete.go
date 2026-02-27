package apikeys

import (
	"fmt"
	"net/http"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// Delete godoc
//
//	@Summary		Delete an API key
//	@Description	Delete an API key for the authenticated user
//	@Tags			apikeys
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		DeleteAPIKeyRequest	true	"API key ID to delete"
//	@Success		200		{object}	MessageResponse		"API key deleted successfully"
//	@Failure		400		{object}	ErrorResponse		"Invalid request body"
//	@Failure		401		{object}	ErrorResponse		"Unauthorized"
//	@Failure		404		{object}	ErrorResponse		"API key not found"
//	@Failure		500		{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/api-keys [delete]
func (h *APIKeyHandler) Delete(c echo.Context) error {
	user := c.Get("user").(*models.User)

	var req DeleteAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("API key deletion failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.KeyID) == "" {
		logger.Info("API key deletion failed: missing key_id")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: key_id",
		})
	}

	err := models.DeleteAPIKey(h.db.DB(), user.ID, req.KeyID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to delete API key: %v", err))
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "API key not found",
		})
	}

	logger.Info("API key deleted successfully")
	return c.JSON(http.StatusOK, MessageResponse{
		Message: "API key deleted successfully",
	})
}
