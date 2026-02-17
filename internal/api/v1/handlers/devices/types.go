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
}

func NewDeviceHandler(db database.Service) *DeviceHandler {
	return &DeviceHandler{
		db: db,
	}
}

func NewDeviceWebsocketHandler(db database.Service) *DeviceHandler {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	return &DeviceHandler{
		db:        db,
		rabbitURL: &rabbitURL,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ErrorResponse struct {
	Error string `json:"error" example:"message"`
}
