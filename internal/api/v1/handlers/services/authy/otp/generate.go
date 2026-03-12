package otp

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/messagetemplate"
	"interface-api/pkg/rabbitmq"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type queuedMessage struct {
	DeviceID     string `json:"device_id"`
	Contact      string `json:"contact"`
	PlatformName string `json:"platform_name"`
	Text         string `json:"text"`
	Username     string `json:"username"`
}

// Generate godoc
//
//	@Summary		Generate and send OTP
//	@Description	Generate a secure OTP code and send it to the specified identifier via the Authy service
//	@Tags			Services - Authy
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Param			Authorization	header		string				true	"Basic Auth credentials"
//	@Param			X-Matrix-Token	header		string				true	"Matrix token (e.g., mt_xxxxx)"
//	@Param			request			body		GenerateOTPRequest	true	"OTP generation request"
//	@Success		200				{object}	GenerateOTPResponse	"OTP sent successfully"
//	@Failure		400				{object}	ErrorResponse		"Invalid request body or validation error"
//	@Failure		500				{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/services/authy/otp/generate [post]
func (h *Handler) Generate(c echo.Context) error {
	userService, ok := c.Get("user_service").(*models.UserService)
	if !ok {
		logger.Error("User service not found in context")
		return echo.ErrUnauthorized
	}

	user, ok := c.Get("user").(*models.User)
	if !ok {
		logger.Error("User not found in context")
		return echo.ErrUnauthorized
	}

	matrixIdentity, ok := c.Get("matrix_identity").(*models.MatrixIdentity)
	if !ok {
		logger.Error("Matrix identity not found in context")
		return echo.ErrUnauthorized
	}

	var req GenerateOTPRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("OTP generation failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.Identifier) == "" {
		logger.Info("OTP generation failed: missing identifier")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: identifier",
		})
	}

	if strings.TrimSpace(req.Platform) == "" {
		logger.Info("OTP generation failed: missing platform")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: platform",
		})
	}

	if strings.TrimSpace(req.Sender) == "" {
		logger.Info("OTP generation failed: missing sender")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: sender",
		})
	}

	var otpCode string
	var expiresAt time.Time

	txErr := h.db.DB().Transaction(func(tx *gorm.DB) error {
		var err error
		var expiryStr string

		otpCode, expiresAt, err = models.CreateOTP(
			tx, userService.ID, req.Identifier, req.Platform, req.Sender,
		)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create OTP: %v\n%s", err, debug.Stack()))
			return err
		}

		expiryStr = expiresAt.Format("Mon, at 15:04 MST")

		exchangeName := os.Getenv("MESSAGE_EXCHANGE_NAME")
		if exchangeName == "" {
			exchangeName = "shortmesh.messages"
		}

		matrixUsername := matrixIdentity.MatrixUsername
		routingKey := fmt.Sprintf("message.%s.%s", req.Platform, matrixUsername)

		producer, err := rabbitmq.NewProducer(*h.rabbitURL)
		if err != nil {
			logger.Error(fmt.Sprintf("RabbitMQ producer creation failed: %v\n%s", err, debug.Stack()))
			return err
		}
		defer producer.Close()

		if err := producer.DeclareExchange(exchangeName, "topic"); err != nil {
			logger.Error(fmt.Sprintf("RabbitMQ exchange declaration failed: %v\n%s", err, debug.Stack()))
			return err
		}

		messageText := messagetemplate.FormatOTPMessage(otpCode, expiryStr)

		message := queuedMessage{
			DeviceID:     req.Sender,
			Contact:      req.Identifier,
			PlatformName: req.Platform,
			Text:         messageText,
			Username:     matrixUsername,
		}

		if err := producer.Publish(exchangeName, routingKey, message, rabbitmq.DefaultPublishOptions()); err != nil {
			logger.Error(fmt.Sprintf("RabbitMQ message publish failed: %v\n%s", err, debug.Stack()))
			return err
		}

		return nil
	})

	if txErr != nil {
		return echo.ErrInternalServerError
	}

	logger.Info(fmt.Sprintf("OTP sent successfully for user %d (service: %d)", user.ID, userService.ID))
	return c.JSON(http.StatusOK, GenerateOTPResponse{
		Message:   "OTP sent successfully",
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}
