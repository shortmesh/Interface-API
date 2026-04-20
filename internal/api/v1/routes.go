package v1

import (
	"interface-api/internal/api/v1/handlers/devices"
	"interface-api/internal/api/v1/handlers/tokens"
	"interface-api/internal/api/v1/handlers/webhooks"
	"interface-api/internal/database"
	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, db database.Service) {
	tokenHandler := tokens.NewTokenHandler(db)
	deviceWsHandler := devices.NewDeviceWebsocketHandler(db)
	webhookHandler := webhooks.NewWebhookHandler(db)
	bearerAuth := middleware.NewBearerAuth(db)
	basicAuth := middleware.NewBasicAuth()

	// Token routes
	g.POST("/tokens", tokenHandler.Create, basicAuth.Authenticate())

	// Device routes
	g.POST("/devices", deviceWsHandler.Create, bearerAuth.Authenticate())
	g.GET("/devices", deviceWsHandler.List, bearerAuth.Authenticate())
	g.GET("/devices/qr-code", deviceWsHandler.QRCode, bearerAuth.AuthenticateWebSocket())
	g.DELETE("/devices", deviceWsHandler.Delete, bearerAuth.Authenticate())
	g.POST("/devices/:device_id/message", deviceWsHandler.SendMessage, bearerAuth.Authenticate())

	// Webhook routes
	g.POST("/webhooks", webhookHandler.Add, bearerAuth.Authenticate())
	g.GET("/webhooks", webhookHandler.List, bearerAuth.Authenticate())
	g.PUT("/webhooks/:id", webhookHandler.Update, bearerAuth.Authenticate())
	g.DELETE("/webhooks/:id", webhookHandler.Delete, bearerAuth.Authenticate())
}
