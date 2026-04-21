package webhooks

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Add godoc
//
//	@Summary		Add a webhook URL
//	@Description	Register a new webhook URL to receive message notifications for the authenticated user
//	@Tags			webhooks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		AddWebhookRequest	true	"Webhook URL"
//	@Success		201		{object}	WebhookResponse		"Webhook added successfully"
//	@Failure		400		{object}	ErrorResponse		"Invalid request"
//	@Failure		401		{object}	ErrorResponse		"Unauthorized"
//	@Failure		409		{object}	ErrorResponse		"Webhook already exists"
//	@Failure		500		{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/webhooks [post]
func (h *WebhookHandler) Add(c echo.Context) error {
	matrixIdentity, ok := c.Get("matrix_identity").(*models.MatrixIdentity)
	if !ok {
		logger.Error("Matrix identity not found in context")
		return echo.ErrUnauthorized
	}

	var req AddWebhookRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("Webhook creation failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.URL) == "" {
		logger.Info("Webhook creation failed: missing URL")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: url",
		})
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		logger.Info(fmt.Sprintf("Webhook creation failed: invalid URL format - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid URL format",
		})
	}

	targetIdentityID := matrixIdentity.ID

	if !matrixIdentity.IsAdmin {
		adminIdentity, err := models.FindAdminMatrixIdentity(h.db.DB())
		if err == nil && adminIdentity.MatrixUsername == matrixIdentity.MatrixUsername {
			targetIdentityID = adminIdentity.ID
			logger.Debug(fmt.Sprintf("Mapping webhook to host identity ID: %d", adminIdentity.ID))
		}
	}

	webhook, err := models.CreateWebhook(h.db.DB(), targetIdentityID, req.URL)
	if err != nil {
		if err == gorm.ErrDuplicatedKey || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			logger.Info("Webhook creation failed: duplicate URL")
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error: "Webhook URL already exists for this user",
			})
		}
		logger.Error(fmt.Sprintf("Failed to create webhook: %v", err))
		return echo.ErrInternalServerError
	}

	logger.Info("Webhook created successfully")
	return c.JSON(http.StatusCreated, WebhookResponse{
		ID:        webhook.ID,
		URL:       webhook.URL,
		Active:    webhook.Active,
		CreatedAt: webhook.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: webhook.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
