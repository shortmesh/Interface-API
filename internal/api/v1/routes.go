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

	g.POST("/users/register", userHandler.Create)
	g.POST("/users/login", userHandler.Login)
	g.POST("/users/logout", userHandler.Logout, auth.Authenticate())
	g.GET("/stream", deviceWsHandler.Stream, auth.Authenticate())
}
