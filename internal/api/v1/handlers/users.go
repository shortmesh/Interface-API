// Package handlers provides HTTP request handlers for the v1 API endpoints.
package handlers

import (
	"net/http"

	"interface-api/internal/database"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	db database.Service
}

func NewUserHandler(db database.Service) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

func (h *UserHandler) GetUser(c echo.Context) error {
	userID := c.Param("id")
	resp := map[string]string{
		"id":      userID,
		"message": "Get user endpoint",
	}

	return c.JSON(http.StatusOK, resp)
}
