package services

import (
	"net/http"

	"interface-api/internal/database/models"

	"github.com/labstack/echo/v4"
)

// GetServiceStatus godoc
//
//	@Summary		Get service subscription status
//	@Description	Check if the authenticated user has subscribed to a specific service and its status
//	@Tags			services
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Session token in format: Bearer sk_xxxxx"
//	@Param			service_name	path		string					true	"Service name (e.g., 'authy')"
//	@Success		200				{object}	ServiceStatusResponse	"Service status information"
//	@Failure		401				{object}	ErrorResponse			"Unauthorized"
//	@Failure		404				{object}	ErrorResponse			"Service not found or not subscribed"
//	@Failure		500				{object}	ErrorResponse			"Internal server error"
//	@Router			/api/v1/services/{service_name}/status [get]
func (h *ServiceHandler) GetServiceStatus(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return echo.ErrUnauthorized
	}

	serviceName := c.Param("service_name")

	service, err := models.FindServiceByName(h.db.DB(), serviceName)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Service not found",
		})
	}

	userService, err := models.FindUserServiceByUserAndService(h.db.DB(), user.ID, service.ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Not subscribed to this service",
		})
	}

	return c.JSON(http.StatusOK, ServiceStatusResponse{
		Name:         service.Name,
		DisplayName:  service.DisplayName,
		IsSubscribed: true,
		IsEnabled:    userService.IsEnabled,
		IsExpired:    userService.IsExpired(),
	})
}
