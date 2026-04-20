package tokens

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/matrixclient"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Delete godoc
//
//	@Summary		Delete a Matrix token
//	@Description	Delete a Matrix token
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Param			id	path		string			true	"Token ID"
//	@Success		200	{object}	DeleteResponse	"Token deleted successfully"
//	@Failure		400	{object}	ErrorResponse	"Invalid request"
//	@Failure		404	{object}	ErrorResponse	"Token not found"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/tokens/{id} [delete]
func (h *TokenHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		logger.Info("Token deletion failed: ID is required")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "ID is required",
		})
	}

	var identity models.MatrixIdentity
	if err := h.db.DB().First(&identity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info("Token deletion failed: token not found")
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Token not found",
			})
		}
		logger.Error(fmt.Sprintf("Failed to find token: %v", err))
		return echo.ErrInternalServerError
	}

	username := identity.MatrixUsername
	deviceID := identity.MatrixDeviceID

	var count int64
	if err := h.db.DB().Model(&models.MatrixIdentity{}).
		Where("matrix_username = ? AND matrix_device_id = ? AND id != ?", username, deviceID, id).
		Count(&count).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to check for other tokens: %v", err))
		return echo.ErrInternalServerError
	}

	shouldDeleteFromMatrix := count == 0

	if err := h.db.DB().Delete(&models.MatrixIdentity{}, id).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to delete token: %v", err))
		return echo.ErrInternalServerError
	}

	if shouldDeleteFromMatrix {
		matrixClient, err := matrixclient.New()
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to initialize Matrix client for cleanup: %v", err))
		} else {
			deleteReq := &matrixclient.DeleteTokenRequest{
				Username: username,
			}
			_, err = matrixClient.DeleteToken(deleteReq)
			if err != nil {
				logger.Warn(fmt.Sprintf("Failed to delete credentials from Matrix client: %v", err))
			} else {
				logger.Info("Deleted credentials from Matrix client")
			}
		}
	} else {
		logger.Debug(fmt.Sprintf("Token deleted from database only - credentials still in use by %d other token(s)", count))
	}

	logger.Info("Token deleted successfully")
	return c.JSON(http.StatusOK, DeleteResponse{
		Message: "Token deleted successfully",
	})
}
