package adminsession

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"interface-api/internal/database/models"
	"interface-api/internal/middleware"
	"interface-api/pkg/crypto"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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

	if username == "" || password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Username and password required"})
	}

	credential, err := models.FindCredentialByClientID(h.db.DB(), username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("Invalid login credentials")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		}
		logger.Error(fmt.Sprintf("Failed to find credential: %v", err))
		return echo.ErrInternalServerError
	}

	secretHash, err := crypto.Hash(password)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to hash password: %v", err))
		return echo.ErrInternalServerError
	}

	if subtle.ConstantTimeCompare(secretHash, credential.ClientSecret) != 1 {
		logger.Error("Invalid login credentials")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
	}

	if !credential.Active {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
	}

	sessionToken, err := middleware.GenerateSessionToken()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate session token: %v", err))
		return echo.ErrInternalServerError
	}

	expiration := time.Now().Add(2 * time.Hour)
	middleware.StoreSession(sessionToken, expiration, &credential.ID)

	cookie := &http.Cookie{
		Name:     "shortmesh_admin_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, LoginResponse{
		Status: "ok",
		Scopes: credential.Scopes,
	})
}
