package tokens

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"interface-api/internal/api/v1/handlers"
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/masclient"
	"interface-api/pkg/matrixclient"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Create godoc
//
//	@Summary		Create a Matrix token
//	@Description	Create a Matrix identity and get a token for Matrix operations. Use use_host=true to reuse admin credentials, or false to create new credentials.
//	@Tags			tokens,admin
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Security		CookieAuth
//	@Param			request	body		CreateRequest	false	"Token creation options"
//	@Success		201		{object}	CreateResponse	"Matrix token created successfully"
//	@Failure		400		{object}	ErrorResponse	"Invalid request"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/tokens [post]
//	@Router			/api/v1/admin/tokens [post]
func (h *TokenHandler) Create(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("Token creation failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if req.UseHost {
		credential, ok := c.Get("credential").(*models.Credential)
		if ok && credential.Role != models.RoleSuperAdmin {
			return c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "Only super_admin can use use_host=true",
			})
		}
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			logger.Info(fmt.Sprintf("Token creation failed: invalid expires_at format - %v", err))
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid expires_at format. Must be RFC3339 (e.g., 2026-12-31T23:59:59Z)",
			})
		}
		if parsedTime.Before(time.Now().UTC()) {
			logger.Info("Token creation failed: expires_at is in the past")
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "expires_at must be in the future",
			})
		}
		expiresAt = &parsedTime
	}

	var matrixToken string
	txErr := h.db.DB().Transaction(func(tx *gorm.DB) error {
		var username, deviceID string
		var err error

		if req.UseHost {
			adminIdentity, err := models.FindAdminMatrixIdentity(tx)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return &handlers.TxError{
						StatusCode: http.StatusBadRequest,
						Message:    "No admin identity found. Create first token with use_host=false",
					}
				}
				logger.Error(fmt.Sprintf("Failed to find admin identity: %v", err))
				return err
			}
			username = adminIdentity.MatrixUsername
			deviceID = adminIdentity.MatrixDeviceID
			logger.Info("Using host matrix credentials")
		} else {
			masClient, err := masclient.New()
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to initialize MAS client: %v", err))
				return err
			}

			matrixClient, err := matrixclient.New()
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to initialize Matrix client: %v", err))
				return err
			}

			adminToken, err := masClient.GetAdminToken()
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to get MAS admin token:\n%v\n\n%s", err, debug.Stack()))
				return err
			}

			u := uuid.New()
			username = hex.EncodeToString(u[:])[:16]

			userResp, err := masClient.CreateUser(adminToken, username)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to create MAS user:\n%v\n\n%s", err, debug.Stack()))
				return err
			}
			logger.Info("Created MAS user")

			u = uuid.New()
			deviceID = hex.EncodeToString(u[:])[:16]
			sessionResp, err := masClient.CreatePersonalSession(
				adminToken, userResp.Data.ID, deviceID, expiresAt,
			)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to create MAS personal session:\n%v\n\n%s", err, debug.Stack()))
				return err
			}
			matrixAccessToken := sessionResp.Data.Attributes.AccessToken
			logger.Info("Created MAS personal session")

			storeReq := &matrixclient.StoreCredentialsRequest{
				Username:    username,
				AccessToken: matrixAccessToken,
				DeviceID:    deviceID,
			}
			_, err = matrixClient.StoreCredentials(storeReq)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to store Matrix credentials:\n%v\n\n%s", err, debug.Stack()))
				return err
			}
			logger.Info("Stored Matrix credentials")
		}

		var count int64
		if err = tx.Model(&models.MatrixIdentity{}).Count(&count).Error; err != nil {
			return err
		}
		isAdmin := count == 0

		matrixToken, _, err = models.CreateMatrixIdentity(tx, username, deviceID, isAdmin, expiresAt)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create matrix identity: %v", err))
			return err
		}

		if isAdmin {
			logger.Info("Created first token - marked as admin")
		}

		return nil
	})

	if txErr != nil {
		var tErr *handlers.TxError
		if errors.As(txErr, &tErr) {
			return c.JSON(tErr.StatusCode, ErrorResponse{
				Error: tErr.Message,
			})
		}
		return echo.ErrInternalServerError
	}

	logger.Info("Matrix token created successfully")
	return c.JSON(http.StatusCreated, CreateResponse{
		Message: "Matrix token created successfully",
		Token:   matrixToken,
	})
}
