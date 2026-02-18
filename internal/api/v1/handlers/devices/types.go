package devices

import (
	"net/http"
	"os"

	"interface-api/internal/database"

	"github.com/gorilla/websocket"
)

type DeviceHandler struct {
	rabbitURL *string
	db        database.Service
	upgrader  *websocket.Upgrader
}

func NewDeviceHandler(db database.Service) *DeviceHandler {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	return &DeviceHandler{
		db:        db,
		rabbitURL: &rabbitURL,
	}
}

func NewDeviceWebsocketHandler(db database.Service) *DeviceHandler {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &DeviceHandler{
		db:        db,
		rabbitURL: &rabbitURL,
		upgrader:  &upgrader,
	}
}

// CreateDeviceRequest represents the request body for creating a device
type CreateDeviceRequest struct {
	Platform string `json:"platform" example:"wa" validate:"required"`
}

// DeviceResponse represents the response after device operations
type DeviceResponse struct {
	Message   string `json:"message,omitempty" example:"Scan the QR code to link your device"`
	QrCodeURL string `json:"qr_code_url,omitempty" example:"/api/v1/devices/qr-code"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"message"`
}
