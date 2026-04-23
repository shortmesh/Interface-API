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
//	@Description	List all Matrix tokens
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
	var tokens []models.MatrixIdentity
	if err := h.db.DB().Find(&tokens).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to list tokens: %v", err))
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, tokens)
}
