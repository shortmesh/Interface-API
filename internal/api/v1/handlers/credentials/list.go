package credentials

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// List godoc
//
//	@Summary		List credentials
//	@Description	Get all active credentials
//	@Tags			credentials,admin
//	@Produce		json
//	@Security		BasicAuth
//	@Security		CookieAuth
//	@Success		200	{array}		CredentialResponse	"List of credentials"
//	@Failure		403		{object}	ErrorResponse	"Insufficient permissions"
//	@Failure		500	{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/credentials [get]
//	@Router			/api/v1/admin/credentials [get]
func (h *CredentialHandler) List(c echo.Context) error {
	credentials, err := models.FindActiveCredentials(h.db.DB())
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to list credentials: %v", err))
		return echo.ErrInternalServerError
	}

	response := make([]CredentialResponse, len(credentials))
	for i, cred := range credentials {
		response[i] = CredentialResponse{
			ClientID:    cred.ClientID,
			Role:        cred.Role,
			Scopes:      cred.Scopes,
			Description: cred.Description,
			Active:      cred.Active,
			CreatedAt:   cred.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   cred.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.JSON(http.StatusOK, response)
}
