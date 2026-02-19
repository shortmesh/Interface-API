package users

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"runtime/debug"
	"strings"
	"time"

	"interface-api/internal/api/v1/handlers"
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/masclient"
	"interface-api/pkg/matrixclient"
	"interface-api/pkg/password"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Create godoc
// @Summary Create a new user
// @Description Create a new user account.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User creation request"
// @Success 201 {object} UserResponse "User created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body or validation error"
// @Failure 409 {object} ErrorResponse "User with email already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Create(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		logger.Log.Infof("Registration failed: invalid request body - %v", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.Email) == "" {
		logger.Log.Info("Registration failed: missing email")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: email",
		})
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		logger.Log.Info("Registration failed: invalid email format")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid email format",
		})
	}

	if strings.TrimSpace(req.Password) == "" {
		logger.Log.Info("Registration failed: missing password")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: password",
		})
	}

	if err := password.ValidatePassword(req.Password); err != nil {
		logger.Log.Infof("Registration failed: password validation - %v", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Invalid password: %v", err),
		})
	}

	_, err := models.FindUserByEmail(h.db.DB(), req.Email)
	if err == nil {
		logger.Log.Info("Registration failed: email already exists")
		return c.JSON(http.StatusConflict, ErrorResponse{
			Error: "User with this email already exists",
		})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Log.Errorf("Email uniqueness check error: %v", err)
		return echo.ErrInternalServerError
	}

	var sessionToken string
	txErr := h.db.DB().Transaction(func(tx *gorm.DB) error {
		masClient, err := masclient.New()
		if err != nil {
			logger.Log.Errorf("Failed to initialize MAS client: %v", err)
			return err
		}

		matrixClient, err := matrixclient.New()
		if err != nil {
			logger.Log.Errorf("Failed to initialize Matrix client: %v", err)
			return err
		}

		adminToken, err := masClient.GetAdminToken()
		if err != nil {
			logger.Log.Errorf("Failed to get MAS admin token:\n%v\n\n%s", err, debug.Stack())
			return err
		}

		u := uuid.New()
		username := hex.EncodeToString(u[:])[:16]

		userResp, err := masClient.CreateUser(adminToken, username)
		if err != nil {
			logger.Log.Errorf("Failed to create MAS user:\n%v\n\n%s", err, debug.Stack())
			return err
		}
		logger.Log.Info("Created MAS user")

		u = uuid.New()
		deviceID := hex.EncodeToString(u[:])[:16]
		sessionResp, err := masClient.CreatePersonalSession(adminToken, userResp.Data.ID, deviceID)
		if err != nil {
			logger.Log.Errorf("Failed to create MAS personal session:\n%v\n\n%s", err, debug.Stack())
			return err
		}
		matrixAccessToken := sessionResp.Data.Attributes.AccessToken
		logger.Log.Info("Created MAS personal session")

		storeReq := &matrixclient.StoreCredentialsRequest{
			Username:    username,
			AccessToken: matrixAccessToken,
			DeviceID:    deviceID,
		}
		_, err = matrixClient.StoreCredentials(storeReq)
		if err != nil {
			logger.Log.Errorf("Failed to store Matrix credentials:\n%v\n\n%s", err, debug.Stack())
			return err
		}
		logger.Log.Info("Stored Matrix credentials")

		user := &models.User{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err := user.SetEmail(req.Email); err != nil {
			logger.Log.Errorf("Failed to set email:\n%v\n\n%s", err, debug.Stack())
			return err
		}

		if err := user.SetPassword(req.Password); err != nil {
			logger.Log.Errorf("Failed to set password:\n%v\n\n%s", err, debug.Stack())
			return err
		}

		if err := tx.Create(user).Error; err != nil {
			logger.Log.Errorf("Failed to create user: %v", err)
			return err
		}

		matrixProfile := &models.MatrixProfile{
			UserID:    user.ID,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if err := matrixProfile.SetMatrixUsername(username); err != nil {
			logger.Log.Errorf("Failed to set matrix username:\n%v\n\n%s", err, debug.Stack())
			return err
		}

		if err := matrixProfile.SetMatrixDeviceID(deviceID); err != nil {
			logger.Log.Errorf("Failed to set matrix device ID:\n%v\n\n%s", err, debug.Stack())
			return err
		}

		if err := tx.Create(matrixProfile).Error; err != nil {
			logger.Log.Errorf("Failed to create matrix profile: %v", err)
			return err
		}

		sessionToken, err = models.CreateOrUpdateSession(
			tx, user.ID, c.RealIP(), c.Request().UserAgent(),
		)
		if err != nil {
			logger.Log.Errorf("Failed to create session: %v", err)
			return err
		}

		return nil
	})

	if txErr != nil {
		var tErr *handlers.TxError
		if errors.As(txErr, &tErr) {
			return c.JSON(tErr.StatusCode, ErrorResponse{
				Error: tErr.Message,
			})
		}
		return echo.ErrInternalServerError
	}

	logger.Log.Info("User created successfully")
	return c.JSON(http.StatusCreated, UserResponse{
		Message: "User created successfully",
		Token:   sessionToken,
	})
}
