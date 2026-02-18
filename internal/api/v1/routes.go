package v1

import (
	"interface-api/internal/api/v1/handlers/devices"
	"interface-api/internal/api/v1/handlers/users"
	"interface-api/internal/database"
	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, db database.Service) {
	userHandler := users.NewUserHandler(db)
	auth := middleware.NewAuth(db)
	deviceWsHandler := devices.NewDeviceWebsocketHandler(db)

	// Auth routes
	g.POST("/auth/register", userHandler.Create)
	g.POST("/auth/login", userHandler.Login)
	g.POST("/auth/logout", userHandler.Logout, auth.Authenticate())

	// Device routes
	g.POST("/devices", deviceWsHandler.Create, auth.Authenticate())
	g.GET("/devices", deviceWsHandler.List, auth.Authenticate())
	g.GET("/devices/qr-code", deviceWsHandler.QRCode, auth.Authenticate())
	g.DELETE("/devices", deviceWsHandler.Delete, auth.Authenticate())
}
