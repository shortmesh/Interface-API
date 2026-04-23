package adminsession

import (
	"net/http"

	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

// CheckMatrixToken godoc
//
//	@Summary		Check matrix token status
//	@Description	Check if the current admin session has a matrix token attached
//	@Tags			admin
//	@Produce		json
//	@Security		AdminSession
//	@Success		200	{object}	MatrixTokenStatusResponse
//	@Failure		401	{object}	ErrorResponse
//	@Router			/api/v1/admin/matrix-token-status [get]
func (h *AdminSessionHandler) CheckMatrixToken(c echo.Context) error {
	cookie, err := c.Cookie("shortmesh_admin_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Session not found"})
	}

	token := middleware.GetMatrixToken(cookie.Value)
	hasToken := token != ""

	return c.JSON(http.StatusOK, MatrixTokenStatusResponse{HasMatrixToken: hasToken})
}
