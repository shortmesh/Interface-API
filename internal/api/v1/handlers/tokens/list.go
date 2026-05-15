package tokens

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// List godoc
//
//	@Summary		List Matrix tokens
//	@Description	List Matrix tokens
//	@Tags			tokens,admin
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Security		CookieAuth
//	@Success		200	{array}		models.MatrixIdentity	"List of tokens"
//	@Failure		500	{object}	ErrorResponse			"Internal server error"
//	@Router			/api/v1/tokens [get]
//	@Router			/api/v1/admin/tokens [get]
func (h *TokenHandler) List(c echo.Context) error {
	credential, ok := c.Get("credential").(*models.Credential)
	if !ok {
		logger.Error("No credential found in context")
		return echo.ErrUnauthorized
	}

	var tokens []models.MatrixIdentity
	var err error

	if credential.Role == models.RoleSuperAdmin {
		err = h.db.DB().Where("credential_id = ? OR credential_id IS NULL", credential.ID).Find(&tokens).Error
	} else {
		err = h.db.DB().Where("credential_id = ?", credential.ID).Find(&tokens).Error
	}

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to list tokens: %v", err))
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, tokens)
}
