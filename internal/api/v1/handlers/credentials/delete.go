package credentials

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Delete godoc
//
//	@Summary		Delete a credential
//	@Description	Permanently delete a credential by client_id
//	@Tags			credentials,admin
//	@Produce		json
//	@Security		BasicAuth
//	@Security		CookieAuth
//	@Param			client_id	path		string			true	"Client ID"
//	@Success		200			{object}	DeleteResponse	"Credential deleted successfully"
//	@Failure		400			{object}	ErrorResponse	"Invalid request"
//	@Failure		404			{object}	ErrorResponse	"Credential not found"
//	@Failure		403		{object}	ErrorResponse	"Insufficient permissions"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/credentials/{client_id} [delete]
//	@Router			/api/v1/admin/credentials/{client_id} [delete]
func (h *CredentialHandler) Delete(c echo.Context) error {
	clientID := c.Param("client_id")
	if clientID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "client_id is required",
		})
	}

	credential, err := models.FindCredentialByClientID(h.db.DB(), clientID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Credential not found",
			})
		}
		logger.Error(fmt.Sprintf("Failed to find credential: %v", err))
		return echo.ErrInternalServerError
	}

	if err := h.db.DB().Delete(credential).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to delete credential: %v", err))
		return echo.ErrInternalServerError
	}

	logger.Info("Credential deleted successfully")

	return c.JSON(http.StatusOK, DeleteResponse{
		Message: "Credential deleted successfully",
	})
}
