package devices

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/rabbitmq"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type queuedMessage struct {
	DeviceID     string `json:"device_id"`
	Contact      string `json:"contact"`
	PlatformName string `json:"platform_name"`
	Text         string `json:"text"`
	Username     string `json:"username"`
}

// SendMessage godoc
// @Summary Send a message via device
// @Description Queue a message to be sent via the specified device
// @Tags devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param device_id path string true "Device ID"
// @Param request body SendMessageRequest true "Message to send"
// @Success 200 {object} SendMessageResponse "Message queued successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/devices/{device_id}/message [post]
func (h *DeviceHandler) SendMessage(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		logger.Log.Error("User not found in context")
		return echo.ErrUnauthorized
	}

	deviceID := c.Param("device_id")
	if strings.TrimSpace(deviceID) == "" {
		logger.Log.Info("Message send failed: missing device_id")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "device_id is required",
		})
	}

	for _, ch := range deviceID {
		if ch < '0' || ch > '9' {
			logger.Log.Info("Message send failed: device_id must be numeric")
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "device_id must contain only numbers",
			})
		}
	}

	var req SendMessageRequest
	if err := c.Bind(&req); err != nil {
		logger.Log.Infof("Message send failed: invalid request body - %v", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.Contact) == "" {
		logger.Log.Info("Message send failed: missing contact")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: contact",
		})
	}

	if strings.TrimSpace(req.Platform) == "" {
		logger.Log.Info("Message send failed: missing platform")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: platform",
		})
	}

	if strings.TrimSpace(req.Text) == "" {
		logger.Log.Info("Message send failed: missing text")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: text",
		})
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

	exchangeName := os.Getenv("MESSAGE_EXCHANGE_NAME")
	if exchangeName == "" {
		exchangeName = "shortmesh.messages"
	}

	routingKey := fmt.Sprintf("message.%s.%s", req.Platform, matrixUsername)

	producer, err := rabbitmq.NewProducer(*h.rabbitURL)
	if err != nil {
		logger.Log.Errorf("RabbitMQ producer creation failed: %v\n%s", err, debug.Stack())
		return echo.ErrInternalServerError
	}
	defer producer.Close()

	if err := producer.DeclareExchange(exchangeName, "topic"); err != nil {
		logger.Log.Errorf("RabbitMQ exchange declaration failed: %v\n%s", err, debug.Stack())
		return echo.ErrInternalServerError
	}

	message := queuedMessage{
		DeviceID:     deviceID,
		Contact:      req.Contact,
		PlatformName: req.Platform,
		Text:         req.Text,
		Username:     matrixUsername,
	}

	if err := producer.Publish(exchangeName, routingKey, message, rabbitmq.DefaultPublishOptions()); err != nil {
		logger.Log.Errorf("RabbitMQ message publish failed: %v\n%s", err, debug.Stack())
		return echo.ErrInternalServerError
	}

	logger.Log.Info("Message queued successfully")
	return c.JSON(http.StatusOK, SendMessageResponse{
		Message: "Message queued successfully",
	})
}
