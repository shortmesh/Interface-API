package adminsession

import (
	"fmt"
	"net/http"

	"interface-api/internal/database/models"
	"interface-api/internal/middleware"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// SetMatrixToken godoc
//
//	@Summary		Set matrix token for admin session
//	@Description	Attach a matrix token to the current admin session for device/webhook operations
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		CookieAuth
//	@Param			request	body		SetMatrixTokenRequest	true	"Matrix token"
//	@Success		200		{object}	SetMatrixTokenResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/admin/matrix-token [post]
func (h *AdminSessionHandler) SetMatrixToken(c echo.Context) error {
	cookie, err := c.Cookie("shortmesh_admin_token")
	if err != nil {
		logger.Info("Set matrix token failed: session not found")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Session not found"})
	}

	var req SetMatrixTokenRequest
	if err := c.Bind(&req); err != nil {
		logger.Info("Set matrix token failed: invalid request body")
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if req.Token == "" {
		logger.Info("Set matrix token failed: token is required")
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Token is required"})
	}

	_, err = models.FindMatrixIdentityByToken(h.db.DB(), req.Token)
	if err != nil {
		logger.Info("Set matrix token failed: invalid token")
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid matrix token"})
	}

	if err := middleware.SetMatrixToken(cookie.Value, req.Token); err != nil {
		logger.Error(fmt.Sprintf("Failed to set matrix token: %v", err))
		return echo.ErrInternalServerError
	}

	logger.Info("Matrix token set for admin session")
	return c.JSON(http.StatusOK, SetMatrixTokenResponse{
		Status:  "ok",
		Message: "Matrix token attached to session",
	})
}
