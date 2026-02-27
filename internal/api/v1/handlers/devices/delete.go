package devices

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/matrixclient"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Delete godoc
//
//	@Summary		Delete a device
//	@Description	Delete a device for the authenticated user
//	@Tags			devices
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		DeleteDeviceRequest	true	"Device deletion request"
//	@Success		200		{object}	DeviceResponse		"Device deleted successfully"
//	@Failure		400		{object}	ErrorResponse		"Invalid request body or validation error"
//	@Failure		500		{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/devices [delete]
func (h *DeviceHandler) Delete(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		logger.Error("User not found in context")
		return echo.ErrUnauthorized
	}

	var req DeleteDeviceRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("Device deletion failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.DeviceID) == "" {
		logger.Info("Device deletion failed: missing device_id")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: device_id",
		})
	}

	if strings.TrimSpace(req.Platform) == "" {
		logger.Info("Device deletion failed: missing platform")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: platform",
		})
	}

	matrixProfile, err := models.FindMatrixProfileByUserID(h.db.DB(), user.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("Matrix profile not found for user")
			return echo.ErrUnauthorized
		}
		logger.Error(fmt.Sprintf("Matrix profile lookup error: %v", err))
		return echo.ErrInternalServerError
	}

	matrixUsername, err := matrixProfile.GetMatrixUsername()
	if err != nil {
		logger.Error(fmt.Sprintf("Matrix username decryption failed: %v", err))
		return echo.ErrInternalServerError
	}

	matrixClient, err := matrixclient.New()
	if err != nil {
		return echo.ErrInternalServerError
	}

	deleteDeviceReq := &matrixclient.DeleteDeviceRequest{
		Username:     matrixUsername,
		DeviceID:     req.DeviceID,
		PlatformName: req.Platform,
	}
	_, err = matrixClient.DeleteDevice(deleteDeviceReq)
	if err != nil {
		logger.Error(fmt.Sprintf("Matrix device deletion failed: %v\n%s", err, debug.Stack()))
		return echo.ErrInternalServerError
	}

	logger.Info("Device deleted successfully")
	return c.JSON(http.StatusOK, DeviceResponse{
		Message: "Device deleted successfully",
	})
}
