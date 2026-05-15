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

// Create godoc
//
//	@Summary		Create a credential
//	@Description	Create a new API credential with auto-generated client_id and client_secret
//	@Tags			credentials,admin
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Security		CookieAuth
//	@Param			request	body		CreateRequest	false	"Credential details"
//	@Success		201		{object}	CreateResponse	"Credential created successfully"
//	@Failure		400		{object}	ErrorResponse	"Invalid request"
//	@Failure		403		{object}	ErrorResponse	"Insufficient permissions"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/credentials [post]
//	@Router			/api/v1/admin/credentials [post]
func (h *CredentialHandler) Create(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("Credential creation failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	const maxRetries = 5
	var clientID string
	var err error

	for i := range maxRetries {
		clientID, err = crypto.GenerateSecureToken(16)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to generate client ID: %v", err))
			return echo.ErrInternalServerError
		}

		_, err = models.FindCredentialByClientID(h.db.DB(), clientID)
		if err == gorm.ErrRecordNotFound {
			break
		}
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to check existing credential: %v", err))
			return echo.ErrInternalServerError
		}

		if i == maxRetries-1 {
			logger.Error("Failed to generate unique client ID after max retries")
			return echo.ErrInternalServerError
		}
	}

	clientSecret, err := crypto.GenerateSecureToken(32)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate client secret: %v", err))
		return echo.ErrInternalServerError
	}

	secretHash, err := crypto.Hash(clientSecret)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to hash client secret: %v", err))
		return echo.ErrInternalServerError
	}

	credential, err := models.UpsertCredential(
		h.db.DB(),
		clientID,
		secretHash,
		models.RoleUser,
		req.Description,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create credential: %v", err))
		return echo.ErrInternalServerError
	}

	logger.Info("Credential created successfully")

	return c.JSON(http.StatusCreated, CreateResponse{
		Message: "Credential created successfully",
		Credential: &CredentialResponse{
			ClientID:    credential.ClientID,
			Role:        credential.Role,
			Scopes:      credential.Scopes,
			Description: credential.Description,
			Active:      credential.Active,
			CreatedAt:   credential.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   credential.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		ClientSecret: clientSecret,
	})
}
