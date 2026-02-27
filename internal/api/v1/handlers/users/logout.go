package users

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// Logout godoc
//
//	@Summary		User logout
//	@Description	Invalidate the current session token
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	UserResponse	"Logout successful"
//	@Failure		401	{object}	ErrorResponse	"Unauthorized"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/auth/logout [post]
func (h *UserHandler) Logout(c echo.Context) error {
	session, ok := c.Get("session").(*models.Session)
	if !ok {
		logger.Error("Session not found in context during logout")
		return echo.ErrUnauthorized
	}

	if err := h.db.DB().Delete(session).Error; err != nil {
		logger.Error(fmt.Sprintf("Session deletion failed: %v", err))
		return echo.ErrInternalServerError
	}

	logger.Info("User logged out successfully")
	return c.JSON(http.StatusOK, UserResponse{
		Message: "Logout successful",
	})
}
