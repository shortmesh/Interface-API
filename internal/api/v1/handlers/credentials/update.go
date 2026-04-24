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
//	@Description	Update credential properties (regenerate secret, activate/deactivate, or update description)
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
//	@Failure		403			{object}	ErrorResponse	"Insufficient permissions"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/credentials/{client_id} [put]
//	@Router			/api/v1/admin/credentials/{client_id} [put]
func (h *CredentialHandler) Update(c echo.Context) error {
	clientID := c.Param("client_id")
	if clientID == "" {
		logger.Info("Credential update failed: missing client_id")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "client_id is required",
		})
	}

	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("Credential update failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
		})
	}

	credential, err := models.FindCredentialByClientID(h.db.DB(), clientID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info(fmt.Sprintf("Credential update failed: client_id '%s' not found", clientID))
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Credential not found",
			})
		}
		logger.Error(fmt.Sprintf("Failed to find credential: %v", err))
		return echo.ErrInternalServerError
	}

	var newSecret *string

	if req.Active != nil {
		if credential.Role == models.RoleSuperAdmin && !*req.Active {
			logger.Info(fmt.Sprintf("Credential update failed: cannot deactivate super admin '%s'", clientID))
			return c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "Cannot deactivate super admin credentials",
			})
		}
		credential.Active = *req.Active
	}

	if req.RegenerateSecret != nil && *req.RegenerateSecret {
		if credential.Role == models.RoleSuperAdmin {
			logger.Info(fmt.Sprintf("Credential update failed: cannot regenerate super admin secret '%s'", clientID))
			return c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "Cannot regenerate secret for super admin credentials",
			})
		}

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
