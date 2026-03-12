package services

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/crypto"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// SubscribeToService godoc
//
//	@Summary		Subscribe to a service
//	@Description	Subscribe the authenticated user to a service with auto-generated credentials
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string				true	"Session token in format: Bearer sk_xxxxx"
//	@Param			service_name	path		string				true	"Service name (e.g., 'authy')"
//	@Success		200				{object}	SubscribeResponse	"Service subscribed successfully"
//	@Failure		401				{object}	ErrorResponse		"Unauthorized"
//	@Failure		404				{object}	ErrorResponse		"Service not found"
//	@Failure		500				{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/services/{service_name}/subscribe [post]
func (h *ServiceHandler) SubscribeToService(c echo.Context) error {
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

	clientID, err := crypto.GenerateSecureToken(16)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate client ID: %v", err))
		return echo.ErrInternalServerError
	}

	clientSecret, err := crypto.GenerateSecureToken(32)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate client secret: %v", err))
		return echo.ErrInternalServerError
	}

	userService, err := models.CreateOrUpdateUserService(h.db.DB(), user.ID, service.ID, clientID, clientSecret, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to subscribe to service: %v", err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, SubscribeResponse{
		Message:      "Successfully subscribed to service",
		ServiceName:  service.Name,
		ClientID:     userService.ClientID,
		ClientSecret: userService.ClientSecret,
	})
}
