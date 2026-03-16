package v1

import (
	"interface-api/internal/api/v1/handlers/devices"
	"interface-api/internal/api/v1/handlers/tokens"
	"interface-api/internal/database"
	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, db database.Service) {
	tokenHandler := tokens.NewTokenHandler(db)
	deviceWsHandler := devices.NewDeviceWebsocketHandler(db)
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
}
