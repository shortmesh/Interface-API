package v1

import (
	"interface-api/internal/api/v1/handlers"
	"interface-api/internal/database"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, db database.Service) {
	userHandler := handlers.NewUserHandler(db)

	g.POST("/users", userHandler.CreateUser)
}
