package devices

import (
	"context"
	"fmt"
	"runtime/debug"

	"interface-api/internal/database/models"
	"interface-api/internal/logger"
	"interface-api/pkg/rabbitmq"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// QRCode godoc
// @Summary WebSocket qr-code endpoint (Not executable in Swagger UI)
// @Description Establishes a WebSocket connection to stream real-time add devices qr-code. This endpoint cannot be tested in Swagger UI - use a WebSocket client instead.
// @Tags devices
// @Security BearerAuth
// @Produce json
// @Success 101 {string} string "WebSocket connection established"
// @Failure 401 {object} ErrorResponse "Missing or invalid authentication token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/devices/qr-code [get]
// @deprecated
func (h *DeviceHandler) QRCode(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		logger.Log.Error("Failed to get user from context")
		return echo.ErrUnauthorized
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

	queueName := matrixUsername

	ws, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.Log.Errorf("Failed to upgrade connection:\n%v\n\n%s", err, debug.Stack())
		return err
	}
	defer ws.Close()

	logger.Log.Info("WebSocket connection established.")

	consumer, err := rabbitmq.NewConsumer(*h.rabbitURL)
	if err != nil {
		logger.Log.Errorf("Failed to create RabbitMQ consumer:\n%v\n\n%s", err, debug.Stack())
		ws.WriteMessage(
			websocket.TextMessage,
			fmt.Append(nil, "Error: Oops, something went wrong. Please try again later."),
		)
		return err
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageChan := make(chan []byte, 100)

	messageHandler := func(body []byte) error {
		select {
		case messageChan <- body:
			return nil
		default:
			logger.Log.Warn("Message channel full, dropping message")
			return nil
		}
	}

	err = consumer.ConsumeQueue(ctx, queueName, messageHandler)
	if err != nil {
		logger.Log.Errorf("Failed to start consuming qr-code queue:\n%v\n\n%s", err, debug.Stack())
		ws.WriteMessage(
			websocket.TextMessage,
			fmt.Append(nil, "Error: You have no pending devices to add. Add a device and try again."),
		)
		return err
	}

	logger.Log.Info("Started consuming from qr-code queue")

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				logger.Log.Infof("WebSocket read error (client disconnected): %v", err)
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			logger.Log.Info("WebSocket connection closed for user")
			return nil
		case msg := <-messageChan:
			err := ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logger.Log.Errorf("Failed to write message to WebSocket: %v", err)
				return err
			}
			logger.Log.Debugf("Sent message to user %s: %s", matrixUsername, string(msg))
		case <-ctx.Done():
			logger.Log.Info("Context cancelled for user")
			return nil
		}
	}
}
