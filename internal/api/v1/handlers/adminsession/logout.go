package adminsession

import (
	"net/http"

	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

// Logout godoc
//
//	@Summary		Admin logout
//	@Description	Clear admin session and redirect to login
//	@Tags			admin
//	@Produce		json
//	@Security		AdminSession
//	@Success		302	"Redirect to login page"
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
	return c.Redirect(http.StatusFound, "/admin/login")
}
