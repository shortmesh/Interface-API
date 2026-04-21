package webhooks

import (
	"fmt"
	"net/http"
	"strconv"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// Delete godoc
//
//	@Summary		Delete a webhook
//	@Description	Remove a webhook for the authenticated user
//	@Tags			webhooks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int				true	"Webhook ID"
//	@Success		200	{object}	MessageResponse	"Webhook deleted successfully"
//	@Failure		400	{object}	ErrorResponse	"Invalid ID"
//	@Failure		401	{object}	ErrorResponse	"Unauthorized"
//	@Failure		404	{object}	ErrorResponse	"Webhook not found"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/webhooks/{id} [delete]
func (h *WebhookHandler) Delete(c echo.Context) error {
	matrixIdentity, ok := c.Get("matrix_identity").(*models.MatrixIdentity)
	if !ok {
		logger.Error("Matrix identity not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Info(fmt.Sprintf("Webhook deletion failed: invalid ID - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid webhook ID",
		})
	}

	targetIdentityID := matrixIdentity.ID

	if !matrixIdentity.IsAdmin {
		adminIdentity, err := models.FindAdminMatrixIdentity(h.db.DB())
		if err == nil && adminIdentity.MatrixUsername == matrixIdentity.MatrixUsername {
			targetIdentityID = adminIdentity.ID
		}
	}

	err = models.DeleteWebhook(h.db.DB(), targetIdentityID, uint(id))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to delete webhook: %v", err))
		return echo.ErrInternalServerError
	}

	logger.Info("Webhook deleted successfully")
	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Webhook deleted successfully",
	})
}
