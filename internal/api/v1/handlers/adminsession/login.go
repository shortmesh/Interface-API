package adminsession

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"time"

	"interface-api/internal/middleware"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

// Login godoc
//
//	@Summary		Admin login
//	@Description	Authenticate admin user and create session
//	@Tags			admin
//	@Accept			x-www-form-urlencoded
//	@Produce		json
//	@Param			username	formData	string	true	"Admin username"
//	@Param			password	formData	string	true	"Admin password"
//	@Success		200			{object}	LoginResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/api/v1/admin/login [post]
func (h *AdminSessionHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Admin credentials not configured"})
	}

	if subtle.ConstantTimeCompare([]byte(username), []byte(clientID)) != 1 ||
		subtle.ConstantTimeCompare([]byte(password), []byte(clientSecret)) != 1 {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
	}

	sessionToken, err := middleware.GenerateSessionToken()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate session token: %v", err))
		return echo.ErrInternalServerError
	}

	expiration := time.Now().Add(2 * time.Hour)
	middleware.StoreSession(sessionToken, expiration)

	cookie := &http.Cookie{
		Name:     "shortmesh_admin_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, LoginResponse{Status: "ok"})
}
