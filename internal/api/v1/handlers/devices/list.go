package devices

import (
	"fmt"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/matrixclient"

	"github.com/labstack/echo/v4"
)

// List godoc
//
//	@Summary		List all devices
//	@Description	List all devices for the Matrix identity
//	@Tags			devices
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string	false	"Matrix token in format: Bearer mt_xxxxx (obtained from /tokens)"
//	@Security		BearerAuth
//	@Success		200	{array}		Device			"List of devices"
//	@Failure		401	{object}	ErrorResponse	"Invalid or expired matrix token"
//	@Failure		403	{object}	ErrorResponse	"Invalid or expired matrix token"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/devices [get]
func (h *DeviceHandler) List(c echo.Context) error {
	matrixIdentity, ok := c.Get("matrix_identity").(*models.MatrixIdentity)
	if !ok {
		logger.Error("Matrix identity not found in context")
		return echo.ErrUnauthorized
	}

	matrixUsername := matrixIdentity.MatrixUsername

	matrixClient, err := matrixclient.New()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create Matrix client: %v", err))
		return echo.ErrInternalServerError
	}

	listDevicesReq := &matrixclient.ListDevicesRequest{
		Username: matrixUsername,
	}
	devices, err := matrixClient.ListDevices(listDevicesReq)
	if err != nil {
		logger.Error(fmt.Sprintf("Matrix device list retrieval failed: %v", err))
		return echo.ErrInternalServerError
	}

	response := make(ListDevicesResponse, 0, len(devices))
	for _, device := range devices {
		response = append(response, Device{
			Platform: device.BridgeName,
			DeviceID: device.DeviceID,
		})
	}

	return c.JSON(200, response)
}
