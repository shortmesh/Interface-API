package services

import (
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// ListUserServices godoc
//
//	@Summary		List user's subscribed services
//	@Description	Get all services the authenticated user has subscribed to, including their status
//	@Tags			services
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string			true	"Session token in format: Bearer sk_xxxxx"
//	@Success		200				{array}		UserServiceInfo	"List of user's services"
//	@Failure		401				{object}	ErrorResponse	"Unauthorized"
//	@Failure		500				{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/services/subscriptions [get]
func (h *ServiceHandler) ListUserServices(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return echo.ErrUnauthorized
	}

	userServices, err := models.FindUserServicesByUser(h.db.DB(), user.ID)
	if err != nil {
		logger.Error("Failed to fetch user services: " + err.Error())
		return echo.ErrInternalServerError
	}

	serviceList := make([]UserServiceInfo, len(userServices))
	for i, us := range userServices {
		var expiresAt *string
		if us.ExpiresAt != nil {
			exp := us.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
			expiresAt = &exp
		}

		serviceList[i] = UserServiceInfo{
			Name:         us.Service.Name,
			DisplayName:  us.Service.DisplayName,
			Description:  us.Service.Description,
			IsEnabled:    us.IsEnabled,
			IsExpired:    us.IsExpired(),
			ClientID:     us.ClientID,
			ClientSecret: us.ClientSecret,
			CreatedAt:    us.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt:    expiresAt,
		}
	}

	return c.JSON(http.StatusOK, serviceList)
}
