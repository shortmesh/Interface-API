package credentials

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/crypto"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Update godoc
//
//	@Summary		Update a credential
//	@Description	Update credential properties (regenerate secret, deactivate, or update description)
//	@Tags			credentials,admin
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Security		CookieAuth
//	@Param			client_id	path		string			true	"Client ID"
//	@Param			request		body		UpdateRequest	true	"Update options"
//	@Success		200			{object}	UpdateResponse	"Credential updated successfully"
//	@Failure		400			{object}	ErrorResponse	"Invalid request"
//	@Failure		404			{object}	ErrorResponse	"Credential not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/credentials/{client_id} [put]
//	@Router			/api/v1/admin/credentials/{client_id} [put]
func (h *CredentialHandler) Update(c echo.Context) error {
	clientID := c.Param("client_id")
	if clientID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "client_id is required",
		})
	}

	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
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

	var newSecret *string

	if req.Deactivate != nil && *req.Deactivate {
		credential.Active = false
	}

	if req.RegenerateSecret != nil && *req.RegenerateSecret {
		generatedSecret, err := crypto.GenerateSecureToken(32)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to generate secret: %v", err))
			return echo.ErrInternalServerError
		}

		secretHash, err := crypto.Hash(generatedSecret)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to hash secret: %v", err))
			return echo.ErrInternalServerError
		}

		credential.ClientSecret = secretHash
		newSecret = &generatedSecret
	}

	if req.Description != nil {
		credential.Description = *req.Description
	}

	if err := h.db.DB().Save(credential).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to update credential: %v", err))
		return echo.ErrInternalServerError
	}

	logger.Info("Credential updated successfully")

	return c.JSON(http.StatusOK, UpdateResponse{
		Message: "Credential updated successfully",
		Credential: &CredentialResponse{
			ClientID:    credential.ClientID,
			Role:        credential.Role,
			Scopes:      credential.Scopes,
			Description: credential.Description,
			Active:      credential.Active,
			CreatedAt:   credential.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   credential.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		ClientSecret: newSecret,
	})
}
