package v1

import (
	"interface-api/internal/api/v1/handlers/devices"
	"interface-api/internal/api/v1/handlers/tokens"
	"interface-api/internal/api/v1/handlers/users"
	"interface-api/internal/database"
	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, db database.Service) {
	userHandler := users.NewUserHandler(db)
	tokenHandler := tokens.NewTokenHandler(db)
	auth := middleware.NewAuth(db)
	deviceWsHandler := devices.NewDeviceWebsocketHandler(db)

	// Auth routes
	g.POST("/auth/register", userHandler.Create)
	g.POST("/auth/login", userHandler.Login)
	g.POST("/auth/logout", userHandler.Logout, auth.Authenticate(middleware.AuthMethodSession))

	// Token routes
	g.POST("/tokens", tokenHandler.Create, auth.Authenticate(middleware.AuthMethodSession))

	// Device routes
	g.POST("/devices", deviceWsHandler.Create, auth.Authenticate(middleware.AuthMethodMatrixToken))
	g.GET("/devices", deviceWsHandler.List, auth.Authenticate(middleware.AuthMethodMatrixToken))
	g.GET("/devices/qr-code", deviceWsHandler.QRCode, auth.Authenticate(middleware.AuthMethodMatrixToken))
	g.DELETE("/devices", deviceWsHandler.Delete, auth.Authenticate(middleware.AuthMethodMatrixToken))
	g.POST("/devices/:device_id/message", deviceWsHandler.SendMessage, auth.Authenticate(middleware.AuthMethodMatrixToken))
}
