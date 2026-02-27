package devices

import (
	"context"
	"fmt"
	"runtime/debug"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/rabbitmq"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

// QRCode godoc
//
//	@Summary		WebSocket qr-code endpoint (Not executable in Swagger UI)
//	@Description	Establishes a WebSocket connection to stream real-time add devices qr-code. This endpoint cannot be tested in Swagger UI - use a WebSocket client instead.
//	@Tags			devices
//	@Security		BearerAuth
//	@Produce		json
//	@Success		101	{string}	string			"WebSocket connection established"
//	@Failure		401	{object}	ErrorResponse	"Missing or invalid authentication token"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/devices/qr-code [get]
//	@deprecated
func (h *DeviceHandler) QRCode(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		logger.Error("User not found in context")
		return echo.ErrUnauthorized
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

	queueName := matrixUsername

	ws, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.Error(fmt.Sprintf("WebSocket upgrade failed: %v\n%s", err, debug.Stack()))
		return err
	}
	defer ws.Close()

	logger.Debug("WebSocket connection established")

	consumer, err := rabbitmq.NewConsumer(*h.rabbitURL)
	if err != nil {
		logger.Error(fmt.Sprintf("RabbitMQ consumer creation failed: %v\n%s", err, debug.Stack()))
		ws.WriteMessage(
			websocket.TextMessage,
			fmt.Append(nil, "Error: Oops, something went wrong. Please try again later."),
		)
		return err
	}
	defer consumer.Close()

	messageChan := make(chan []byte, 100)

	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	messageHandler := func(delivery amqp.Delivery) error {
		select {
		case messageChan <- delivery.Body:
			return nil
		case <-ctx.Done():
			return nil
		default:
			logger.Warn("QR code message channel full, dropping message")
			return nil
		}
	}

	err = consumer.Consume(ctx, queueName, messageHandler, cancel, rabbitmq.DefaultConsumeOptions())
	if err != nil {
		logger.Error(fmt.Sprintf("QR code queue consumption failed: %v\n%s", err, debug.Stack()))
		ws.WriteMessage(
			websocket.TextMessage,
			fmt.Append(nil, "Error: You have no pending devices to add. Add a device and try again."),
		)
		return err
	}

	logger.Debug("Started consuming QR code queue")

	done := make(chan struct{})

	go func() {
		defer close(done)
		defer close(messageChan)
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				logger.Debug("WebSocket client disconnected")
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			logger.Debug("WebSocket connection closed")
			return nil
		case msg, ok := <-messageChan:
			if !ok {
				logger.Debug("QR code message channel closed")
				return nil
			}
			err := ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logger.Error(fmt.Sprintf("WebSocket message write failed: %v", err))
				return err
			}
			logger.Debug("QR code sent to client")
		case <-ctx.Done():
			logger.Debug("WebSocket context cancelled")
			return nil
		}
	}
}
