package devices

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/rabbitmq"

	"github.com/labstack/echo/v4"
)

type queuedMessage struct {
	DeviceID      string `json:"device_id"`
	Contact       string `json:"contact"`
	PlatformName  string `json:"platform_name"`
	Text          string `json:"text"`
	Username      string `json:"username"`
	FileContent   string `json:"file_content,omitempty"`
	FileExtension string `json:"file_extension,omitempty"`
}

// SendMessage godoc
//
//	@Summary		Send a message via device
//	@Description	Queue a message to be sent via the specified device. Either text or file must be provided (or both).
//	@Tags			devices
//	@Accept			json,mpfd
//	@Produce		json
//	@Param			Authorization	header	string	false	"Matrix token in format: Bearer mt_xxxxx (obtained from /tokens)"
//	@Security		BearerAuth
//	@Param			device_id	path		string				true	"Device ID"
//	@Param			request		body		SendMessageRequest	false	"Message to send (JSON)"
//	@Param			contact		formData	string				false	"Contact (multipart)"
//	@Param			platform	formData	string				false	"Platform (multipart)"
//	@Param			text		formData	string				false	"Message text (multipart, optional if file provided)"
//	@Param			file		formData	file				false	"File to upload (multipart)"
//	@Success		200			{object}	SendMessageResponse	"Message queued successfully"
//	@Failure		400			{object}	ErrorResponse		"Invalid request body or validation error"
//	@Failure		401			{object}	ErrorResponse		"Invalid or expired matrix token"
//	@Failure		403			{object}	ErrorResponse		"Invalid or expired matrix token"
//	@Failure		500			{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/devices/{device_id}/message [post]
func (h *DeviceHandler) SendMessage(c echo.Context) error {
	matrixIdentity, ok := c.Get("matrix_identity").(*models.MatrixIdentity)
	if !ok {
		logger.Error("Matrix identity not found in context")
		return echo.ErrUnauthorized
	}

	deviceID := c.Param("device_id")
	if strings.TrimSpace(deviceID) == "" {
		logger.Info("Message send failed: missing device_id")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "device_id is required",
		})
	}

	var req SendMessageRequest
	var fileContent string
	var fileExtension string

	contentType := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := c.Request().ParseMultipartForm(32 << 20); err != nil { // 32 MB max
			logger.Info(fmt.Sprintf("Message send failed: cannot parse multipart form - %v", err))
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid multipart form data",
			})
		}

		req.Contact = c.FormValue("contact")
		req.Platform = c.FormValue("platform")
		req.Text = c.FormValue("text")

		file, err := c.FormFile("file")
		if err == nil && file != nil {
			src, err := file.Open()
			if err != nil {
				logger.Info(fmt.Sprintf("Message send failed: cannot open uploaded file - %v", err))
				return c.JSON(http.StatusBadRequest, ErrorResponse{
					Error: "Failed to process uploaded file",
				})
			}
			defer src.Close()

			fileBytes, err := io.ReadAll(src)
			if err != nil {
				logger.Info(fmt.Sprintf("Message send failed: cannot read uploaded file - %v", err))
				return c.JSON(http.StatusBadRequest, ErrorResponse{
					Error: "Failed to read uploaded file",
				})
			}

			fileContent = base64.StdEncoding.EncodeToString(fileBytes)
			fileExtension = strings.TrimPrefix(filepath.Ext(file.Filename), ".")
			if fileExtension == "" {
				logger.Info("Message send failed: uploaded file has no extension")
				return c.JSON(http.StatusBadRequest, ErrorResponse{
					Error: "Uploaded file must have a file extension",
				})
			}
		}
	} else {
		if err := c.Bind(&req); err != nil {
			logger.Info(fmt.Sprintf("Message send failed: invalid request body - %v", err))
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid request body. Must be a JSON object.",
			})
		}
	}

	if strings.TrimSpace(req.Contact) == "" {
		logger.Info("Message send failed: missing contact")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: contact",
		})
	}

	if strings.TrimSpace(req.Platform) == "" {
		logger.Info("Message send failed: missing platform")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: platform",
		})
	}

	if strings.TrimSpace(req.Text) == "" && fileContent == "" {
		logger.Info("Message send failed: missing text and file")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Either text or file must be provided",
		})
	}

	matrixUsername := matrixIdentity.MatrixUsername

	exchangeName := os.Getenv("MESSAGE_EXCHANGE_NAME")
	if exchangeName == "" {
		exchangeName = "shortmesh.messages"
	}

	routingKey := fmt.Sprintf("message.%s.%s", req.Platform, matrixUsername)

	producer, err := rabbitmq.NewProducer(*h.rabbitURL)
	if err != nil {
		logger.Error(fmt.Sprintf("RabbitMQ producer creation failed: %v\n%s", err, debug.Stack()))
		return echo.ErrInternalServerError
	}
	defer producer.Close()

	if err := producer.DeclareExchange(exchangeName, "topic"); err != nil {
		logger.Error(fmt.Sprintf("RabbitMQ exchange declaration failed: %v\n%s", err, debug.Stack()))
		return echo.ErrInternalServerError
	}

	message := queuedMessage{
		DeviceID:      deviceID,
		Contact:       req.Contact,
		PlatformName:  req.Platform,
		Text:          req.Text,
		Username:      matrixUsername,
		FileContent:   fileContent,
		FileExtension: fileExtension,
	}

	if err := producer.Publish(exchangeName, routingKey, message, rabbitmq.DefaultPublishOptions()); err != nil {
		logger.Error(fmt.Sprintf("RabbitMQ message publish failed: %v\n%s", err, debug.Stack()))
		return echo.ErrInternalServerError
	}

	logger.Info("Message queued successfully")
	return c.JSON(http.StatusOK, SendMessageResponse{
		Message: "Message queued successfully",
	})
}
