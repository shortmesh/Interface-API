package users

import (
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/internal/logger"

	"github.com/labstack/echo/v4"
)

// Logout godoc
// @Summary User logout
// @Description Invalidate the current session token
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse "Logout successful"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/users/logout [post]
func (h *UserHandler) Logout(c echo.Context) error {
	session, ok := c.Get("session").(*models.Session)
	if !ok {
		logger.Log.Error("Failed to get session from context")
		return echo.ErrUnauthorized
	}

	if err := h.db.DB().Delete(session).Error; err != nil {
		logger.Log.Errorf("Failed to delete session: %v", err)
		return echo.ErrInternalServerError
	}

	logger.Log.Info("User logged out successfully")
	return c.JSON(http.StatusOK, UserResponse{
		Message: "Logout successful",
	})
}
