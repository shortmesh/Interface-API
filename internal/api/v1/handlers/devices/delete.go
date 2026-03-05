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
)

// Delete godoc
//
//	@Summary		Delete a device
//	@Description	Delete a device for the Matrix identity
//	@Tags			devices
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string	false	"Matrix token in format: Bearer mt_xxxxx (obtained from /tokens)"
//	@Security		BearerAuth
//	@Param			request	body		DeleteDeviceRequest	true	"Device deletion request"
//	@Success		200		{object}	DeviceResponse		"Device deleted successfully"
//	@Failure		400		{object}	ErrorResponse		"Invalid request body or validation error"
//	@Failure		401		{object}	ErrorResponse		"Invalid or expired matrix token"
//	@Failure		500		{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/devices [delete]
func (h *DeviceHandler) Delete(c echo.Context) error {
	matrixIdentity, ok := c.Get("matrix_identity").(*models.MatrixIdentity)
	if !ok {
		logger.Error("Matrix identity not found in context")
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

	matrixUsername := matrixIdentity.MatrixUsername

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
