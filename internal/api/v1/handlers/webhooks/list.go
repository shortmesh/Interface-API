package webhooks

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// List godoc
//
//	@Summary		List all webhooks
//	@Description	Get a list of all webhooks for the authenticated user
//	@Tags			webhooks
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		WebhookResponse	"List of webhooks"
//	@Failure		401	{object}	ErrorResponse	"Unauthorized"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/webhooks [get]
func (h *WebhookHandler) List(c echo.Context) error {
	matrixIdentity, ok := c.Get("matrix_identity").(*models.MatrixIdentity)
	if !ok {
		logger.Error("Matrix identity not found in context")
		return echo.ErrUnauthorized
	}

	targetIdentityID := matrixIdentity.ID

	if !matrixIdentity.IsAdmin {
		adminIdentity, err := models.FindAdminMatrixIdentity(h.db.DB())
		if err == nil && adminIdentity.MatrixUsername == matrixIdentity.MatrixUsername {
			targetIdentityID = adminIdentity.ID
		}
	}

	webhooks, err := models.FindWebhooksByIdentity(h.db.DB(), targetIdentityID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to fetch webhooks: %v", err))
		return echo.ErrInternalServerError
	}

	response := make([]WebhookResponse, 0)
	for _, webhook := range webhooks {
		response = append(response, WebhookResponse{
			ID:        webhook.ID,
			URL:       webhook.URL,
			Active:    webhook.Active,
			CreatedAt: webhook.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: webhook.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return c.JSON(http.StatusOK, response)
}
