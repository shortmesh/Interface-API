package v1

import (
	"interface-api/internal/api/v1/handlers/apikeys"
	"interface-api/internal/api/v1/handlers/devices"
	"interface-api/internal/api/v1/handlers/users"
	"interface-api/internal/database"
	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, db database.Service) {
	userHandler := users.NewUserHandler(db)
	apiKeyHandler := apikeys.NewAPIKeyHandler(db)
	auth := middleware.NewAuth(db)
	deviceWsHandler := devices.NewDeviceWebsocketHandler(db)

	// Auth routes
	g.POST("/auth/register", userHandler.Create)
	g.POST("/auth/login", userHandler.Login)
	g.POST("/auth/logout", userHandler.Logout, auth.Authenticate())

	// API Key routes
	g.POST("/api-keys", apiKeyHandler.Create, auth.Authenticate(
		middleware.AuthMethodSession, middleware.AuthMethodAPIKey),
	)
	g.GET("/api-keys", apiKeyHandler.List, auth.Authenticate(
		middleware.AuthMethodSession, middleware.AuthMethodAPIKey),
	)
	g.DELETE("/api-keys", apiKeyHandler.Delete, auth.Authenticate(
		middleware.AuthMethodSession, middleware.AuthMethodAPIKey),
	)

	// Device routes
	g.POST("/devices", deviceWsHandler.Create, auth.Authenticate(
		middleware.AuthMethodSession, middleware.AuthMethodAPIKey),
	)
	g.GET("/devices", deviceWsHandler.List, auth.Authenticate(
		middleware.AuthMethodSession, middleware.AuthMethodAPIKey),
	)
	g.GET("/devices/qr-code", deviceWsHandler.QRCode, auth.Authenticate(
		middleware.AuthMethodSession, middleware.AuthMethodAPIKey),
	)
	g.DELETE("/devices", deviceWsHandler.Delete, auth.Authenticate(
		middleware.AuthMethodSession, middleware.AuthMethodAPIKey),
	)
	g.POST("/devices/:device_id/message", deviceWsHandler.SendMessage, auth.Authenticate(
		middleware.AuthMethodSession, middleware.AuthMethodAPIKey),
	)
}
