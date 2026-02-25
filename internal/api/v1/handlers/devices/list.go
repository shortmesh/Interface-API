package devices

import (
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/matrixclient"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// List godoc
//
//	@Summary		List all devices
//	@Description	List all devices for the authenticated user
//	@Tags			devices
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		Device			"List of devices"
//	@Failure		400	{object}	ErrorResponse	"Invalid request body or validation error"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/devices [get]
func (h *DeviceHandler) List(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		logger.Log.Error("User not found in context")
		return echo.ErrUnauthorized
	}

	matrixProfile, err := models.FindMatrixProfileByUserID(h.db.DB(), user.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Log.Warn("Matrix profile not found for user")
			return echo.ErrUnauthorized
		}
		logger.Log.Errorf("Matrix profile lookup error: %v", err)
		return echo.ErrInternalServerError
	}

	matrixUsername, err := matrixProfile.GetMatrixUsername()
	if err != nil {
		logger.Log.Errorf("Matrix username decryption failed: %v", err)
		return echo.ErrInternalServerError
	}

	matrixClient, err := matrixclient.New()
	if err != nil {
		return echo.ErrInternalServerError
	}

	listDevicesReq := &matrixclient.ListDevicesRequest{
		Username: matrixUsername,
	}
	devices, err := matrixClient.ListDevices(listDevicesReq)
	if err != nil {
		logger.Log.Errorf("Matrix device list retrieval failed: %v", err)
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
