package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"interface-api/internal/api/v1/handlers/types"
	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/internal/logger"
	"interface-api/pkg/jwt"
	"interface-api/pkg/password"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type UserHandler struct {
	db database.Service
}

func NewUserHandler(db database.Service) *UserHandler {
	return &UserHandler{db: db}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user account.
// @Tags users
// @Accept json
// @Produce json
// @Param user body types.CreateUserRequest true "User creation request"
// @Success 201 {object} types.UserResponse "User created successfully"
// @Failure 400 {object} types.ErrorResponse "Invalid request body or validation error"
// @Failure 409 {object} types.ErrorResponse "User with email already exists"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req types.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		logger.Log.Errorf("Failed to bind request body: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.Email) == "" {
		logger.Log.Error("Missing required field: email")
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Missing required field: email",
		})
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		logger.Log.Errorf("Invalid email format: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid email format",
		})
	}

	if strings.TrimSpace(req.Password) == "" {
		logger.Log.Error("Missing required field: password")
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Missing required field: password",
		})
	}

	if err := password.ValidatePassword(req.Password); err != nil {
		logger.Log.Errorf("Invalid password: %v", err)
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: fmt.Sprintf("Invalid password: %v", err),
		})
	}

	_, err := models.FindUserByEmail(h.db.DB(), req.Email)
	if err == nil {
		logger.Log.Error("User with email already exists")
		return c.JSON(http.StatusConflict, types.ErrorResponse{
			Error: "User with this email already exists",
		})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Log.Errorf("Failed to check email uniqueness: %v", err)
		return echo.ErrInternalServerError
	}

	var jwtToken string
	err = h.db.DB().Transaction(func(tx *gorm.DB) error {
		user := &models.User{
			PasswordHash: req.Password,
		}

		if err := user.SetEmail(req.Email); err != nil {
			logger.Log.Errorf("Failed to set email: %v", err)
			return err
		}

		if err := tx.Create(user).Error; err != nil {
			logger.Log.Errorf("Failed to create user: %v", err)
			return err
		}

		session, err := models.CreateOrUpdateSession(tx, user.ID, c.RealIP(), c.Request().UserAgent())
		if err != nil {
			logger.Log.Errorf("Failed to create session: %v", err)
			return err
		}

		jwtToken, err = jwt.GenerateToken(user.ID, session.ID, session.Token, session.ExpiresAt)
		if err != nil {
			logger.Log.Errorf("Failed to generate JWT: %v", err)
			return err
		}

		return nil
	})
	if err != nil {
		return echo.ErrInternalServerError
	}

	logger.Log.Info("User created successfully")
	return c.JSON(http.StatusCreated, types.UserResponse{
		Message: "User created successfully",
		Token:   jwtToken,
	})
}
