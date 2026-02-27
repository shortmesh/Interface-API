package users

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticate a user and return a session token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"Login credentials"
//	@Success		200		{object}	UserResponse	"Login successful"
//	@Failure		400		{object}	ErrorResponse	"Invalid request body or validation error"
//	@Failure		401		{object}	ErrorResponse	"Invalid credentials"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/api/v1/auth/login [post]
func (h *UserHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("Login failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.Email) == "" {
		logger.Info("Login failed: missing email")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: email",
		})
	}

	if strings.TrimSpace(req.Password) == "" {
		logger.Info("Login failed: missing password")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: password",
		})
	}

	user, err := models.FindUserByEmail(h.db.DB(), req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("Login failed: invalid credentials (user not found)")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid credentials",
			})
		}
		logger.Error(fmt.Sprintf("User lookup error: %v", err))
		return echo.ErrInternalServerError
	}

	if err := user.ComparePassword(req.Password); err != nil {
		logger.Info("Login failed: invalid credentials (password mismatch)")
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid credentials",
		})
	}

	sessionToken, err := models.CreateOrUpdateSession(
		h.db.DB(), user.ID, c.RealIP(), c.Request().UserAgent(),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Session creation error: %v", err))
		return echo.ErrInternalServerError
	}

	if err := user.RecordLogin(h.db.DB()); err != nil {
		logger.Warn(fmt.Sprintf("Last login update failed: %v", err))
	}

	logger.Info("User logged in successfully")
	return c.JSON(http.StatusOK, UserResponse{
		Message: "Login successful",
		Token:   sessionToken,
	})
}
