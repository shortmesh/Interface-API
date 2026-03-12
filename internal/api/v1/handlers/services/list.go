package services

import (
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// ListAvailableServices godoc
//
//	@Summary		List available services
//	@Description	Get all active services available for subscription
//	@Tags			Services
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Session token in format: Bearer sk_xxxxx"
//	@Success		200				{object}	ListServicesResponse	"List of available services"
//	@Failure		401				{object}	ErrorResponse			"Unauthorized"
//	@Failure		500				{object}	ErrorResponse			"Internal server error"
//	@Router			/api/v1/services [get]
func (h *ServiceHandler) ListAvailableServices(c echo.Context) error {
	services, err := models.FindAllActiveServices(h.db.DB())
	if err != nil {
		logger.Error("Failed to fetch services: " + err.Error())
		return echo.ErrInternalServerError
	}

	serviceList := make([]ServiceInfo, len(services))
	for i, s := range services {
		serviceList[i] = ServiceInfo{
			Name:        s.Name,
			DisplayName: s.DisplayName,
			Description: s.Description,
		}
	}

	return c.JSON(http.StatusOK, ListServicesResponse{
		Services: serviceList,
	})
}
