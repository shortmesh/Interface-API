package devices

import (
	"net/http"
	"runtime/debug"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/internal/logger"
	"interface-api/pkg/matrixclient"
	"interface-api/pkg/rabbitmq"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Create godoc
// @Summary Create a new device
// @Description Create a new device for the authenticated user
// @Tags devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body CreateDeviceRequest true "Device creation request"
// @Success 201 {object} DeviceResponse "Requested to add device successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/devices [post]
func (h *DeviceHandler) Create(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		logger.Log.Error("Failed to get user from context")
		return echo.ErrUnauthorized
	}

	var req CreateDeviceRequest
	if err := c.Bind(&req); err != nil {
		logger.Log.Errorf("Failed to bind request body: %v", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.Platform) == "" {
		logger.Log.Error("Missing required field: platform")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: platform",
		})
	}

	matrixProfile, err := models.FindMatrixProfileByUserID(h.db.DB(), user.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Log.Error("Matrix profile not found for user")
			return echo.ErrUnauthorized
		}
		logger.Log.Errorf("Failed to fetch matrix profile: %v", err)
		return echo.ErrInternalServerError
	}

	matrixUsername, err := matrixProfile.GetMatrixUsername()
	if err != nil {
		logger.Log.Errorf("Failed to decrypt matrix username: %v", err)
		return echo.ErrInternalServerError
	}

	consumer, err := rabbitmq.NewConsumer(*h.rabbitURL)
	if err != nil {
		logger.Log.Errorf("Failed to create RabbitMQ consumer:\n%v\n\n%s", err, debug.Stack())
		return echo.ErrInternalServerError
	}
	defer consumer.Close()

	queueExists, _ := consumer.DoesQueueExist(matrixUsername)

	if queueExists {
		logger.Log.Info("QR-Code already streaming for this user. No need to create a new device.")
		return c.JSON(http.StatusCreated, DeviceResponse{
			Message:   "Scan the QR code to link your device",
			QrCodeURL: "/api/v1/devices/qr-code",
		})
	}

	matrixClient, err := matrixclient.New()
	if err != nil {
		logger.Log.Errorf("Failed to initialize Matrix client: %v", err)
		return echo.ErrInternalServerError
	}

	addDeviceReq := &matrixclient.AddDeviceRequest{
		Username:     matrixUsername,
		PlatformName: req.Platform,
	}
	_, err = matrixClient.AddDevice(addDeviceReq)
	if err != nil {
		logger.Log.Errorf("Failed to request add matrix device:\n%v\n\n%s", err, debug.Stack())
		return echo.ErrInternalServerError
	}
	logger.Log.Info("Requested to add matrix device.")
	return c.JSON(http.StatusCreated, DeviceResponse{
		Message:   "Scan the QR code to link your device",
		QrCodeURL: "/api/v1/devices/qr-code",
	})
}
