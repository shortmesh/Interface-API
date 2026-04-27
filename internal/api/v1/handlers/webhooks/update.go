package webhooks

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Update godoc
//
//	@Summary		Update a webhook
//	@Description	Update webhook URL and/or active status for the authenticated user
//	@Tags			webhooks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"Webhook ID"
//	@Param			request	body		UpdateWebhookRequest	true	"Webhook update data"
//	@Success		200		{object}	WebhookResponse			"Webhook updated successfully"
//	@Failure		400		{object}	ErrorResponse			"Invalid request"
//	@Failure		401		{object}	ErrorResponse			"Unauthorized"
//	@Failure		404		{object}	ErrorResponse			"Webhook not found"
//	@Failure		409		{object}	ErrorResponse			"Webhook URL already exists"
//	@Failure		500		{object}	ErrorResponse			"Internal server error"
//	@Router			/api/v1/webhooks/{id} [put]
func (h *WebhookHandler) Update(c echo.Context) error {
	matrixIdentity, ok := c.Get("matrix_identity").(*models.MatrixIdentity)
	if !ok {
		logger.Error("Matrix identity not found in context")
		return echo.ErrUnauthorized
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Info(fmt.Sprintf("Webhook update failed: invalid ID - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid webhook ID",
		})
	}

	var req UpdateWebhookRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("Webhook update failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if req.URL != nil && strings.TrimSpace(*req.URL) == "" {
		logger.Info("Webhook update failed: empty URL provided")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "URL cannot be empty",
		})
	}

	if req.URL != nil {
		if _, err := url.ParseRequestURI(*req.URL); err != nil {
			logger.Info(fmt.Sprintf("Webhook update failed: invalid URL format - %v", err))
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid URL format",
			})
		}
	}

	targetIdentityID := matrixIdentity.ID

	if !matrixIdentity.IsAdmin {
		adminIdentity, err := models.FindAdminMatrixIdentity(h.db.DB())
		if err == nil && adminIdentity.MatrixUsername == matrixIdentity.MatrixUsername {
			targetIdentityID = adminIdentity.ID
		}
	}

	webhook, err := models.UpdateWebhook(h.db.DB(), targetIdentityID, uint(id), req.URL, req.Active)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info(fmt.Sprintf("Webhook update failed: webhook ID %d not found", id))
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Webhook not found",
			})
		}
		if err == gorm.ErrDuplicatedKey || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			logger.Info("Webhook update failed: URL already exists")
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error: "Webhook URL already exists for this user",
			})
		}
		logger.Error(fmt.Sprintf("Failed to update webhook: %v", err))
		return echo.ErrInternalServerError
	}

	logger.Info("Webhook updated successfully")
	return c.JSON(http.StatusOK, WebhookResponse{
		ID:        webhook.ID,
		URL:       webhook.URL,
		Active:    webhook.Active,
		CreatedAt: webhook.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: webhook.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
