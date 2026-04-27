package adminsession

import (
	"net/http"

	"interface-api/internal/middleware"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// Logout godoc
//
//	@Summary		Admin logout
//	@Description	Clear admin session and redirect to login
//	@Tags			admin
//	@Produce		json
//	@Security		CookieAuth
//	Success		200	{object}	LogoutResponse
//	@Router			/api/v1/admin/logout [get]
func (h *AdminSessionHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("shortmesh_admin_token")
	if err == nil {
		middleware.ClearSession(cookie.Value)
	}

	cookie = &http.Cookie{
		Name:     "shortmesh_admin_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	c.SetCookie(cookie)

	logger.Info("Logged out successfully")

	return c.JSON(http.StatusOK, LogoutResponse{
		Message: "Logged out successfully",
	})
}
