package users

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"runtime/debug"
	"strings"
	"time"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/password"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Create godoc
//
//	@Summary		Create a new user
//	@Description	Create a new user account.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateUserRequest	true	"User creation request"
//	@Success		201		{object}	UserResponse		"User created successfully"
//	@Failure		400		{object}	ErrorResponse		"Invalid request body or validation error"
//	@Failure		409		{object}	ErrorResponse		"User with email already exists"
//	@Failure		500		{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/auth/register [post]
func (h *UserHandler) Create(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("Registration failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.Email) == "" {
		logger.Info("Registration failed: missing email")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: email",
		})
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		logger.Info("Registration failed: invalid email format")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid email format",
		})
	}

	if strings.TrimSpace(req.Password) == "" {
		logger.Info("Registration failed: missing password")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: password",
		})
	}

	if err := password.ValidatePassword(req.Password); err != nil {
		logger.Info(fmt.Sprintf("Registration failed: password validation - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Invalid password: %v", err),
		})
	}

	_, err := models.FindUserByEmail(h.db.DB(), req.Email)
	if err == nil {
		logger.Info("Registration failed: email already exists")
		return c.JSON(http.StatusConflict, ErrorResponse{
			Error: "User with this email already exists",
		})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error(fmt.Sprintf("Email uniqueness check error: %v", err))
		return echo.ErrInternalServerError
	}

	var sessionToken string
	txErr := h.db.DB().Transaction(func(tx *gorm.DB) error {
		user := &models.User{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err := user.SetPassword(req.Password); err != nil {
			logger.Error(fmt.Sprintf("Failed to set password:\n%v\n\n%s", err, debug.Stack()))
			return err
		}

		user.Email = req.Email

		if err := tx.Create(user).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to create user: %v", err))
			return err
		}

		sessionToken, err = models.CreateOrUpdateSession(
			tx, user.ID, c.RealIP(), c.Request().UserAgent(),
		)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create session: %v", err))
			return err
		}

		return nil
	})

	if txErr != nil {
		return echo.ErrInternalServerError
	}

	logger.Info("User created successfully")
	return c.JSON(http.StatusCreated, UserResponse{
		Message: "User created successfully",
		Token:   sessionToken,
	})
}
