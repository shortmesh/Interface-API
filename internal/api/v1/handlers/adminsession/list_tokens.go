package adminsession

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// ListTokens godoc
//
//	@Summary		List matrix tokens
//	@Description	Get all matrix identity tokens (requires admin session)
//	@Tags			admin
//	@Produce		json
//	@Security		AdminSession
//	@Success		200	{array}		models.MatrixIdentity
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/v1/admin/tokens [get]
func (h *AdminSessionHandler) ListTokens(c echo.Context) error {
	var tokens []models.MatrixIdentity
	if err := h.db.DB().Find(&tokens).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to list tokens: %v", err))
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, tokens)
}
