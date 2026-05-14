package devices

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	GroupURL      string `json:"group_url,omitempty"`
	ReplyID       string `json:"reply_id,omitempty"`
}

// SendMessage godoc
//
//	@Summary		Send a message via device
//	@Description	Queue a message to be sent via the specified device. Either text or file must be provided (or both).
//	@Description	At least one of contact or group_url must be provided. If both are provided, group_url takes priority.
//	@Description	The 'contact' field accepts comma-separated phone numbers or group IDs.
//	@Description	The 'group_url' field accepts comma-separated group URLs.
//	@Description	The 'reply_id' field can be used with either contact or group_url to reply to a specific message.
//	@Description	For JSON requests: use file_content (base64) and file_extension to send files.
//	@Description	For multipart requests: use the file field to upload files directly.
//	@Tags			devices
//	@Accept			json,mpfd
//	@Produce		json
//	@Param			Authorization	header	string	false	"Matrix token in format: Bearer mt_xxxxx (obtained from /tokens)"
//	@Security		BearerAuth
//	@Param			device_id	path		string				true	"Device ID"
//	@Param			request		body		SendMessageRequest	false	"Message to send (JSON). Use file_content/file_extension for files in JSON."
//	@Param			contact		formData	string				false	"Phone number or group ID (multipart, optional if group_url or reply_id provided)"
//	@Param			platform	formData	string				false	"Platform (multipart)"
//	@Param			text		formData	string				false	"Message text (multipart, optional if file provided)"
//	@Param			group_url	formData	string				false	"Group URL - use if you don't have group ID (multipart, optional)"
//	@Param			reply_id	formData	string				false	"Reply to message ID (multipart, optional)"
//	@Param			file		formData	file				false	"File to upload (multipart)"
//	@Success		200			{object}	SendMessageResponse	"Message(s) queued successfully"
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
		req.GroupURL = c.FormValue("group_url")
		req.ReplyID = c.FormValue("reply_id")

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

	if strings.TrimSpace(req.Contact) == "" && strings.TrimSpace(req.GroupURL) == "" {
		logger.Info("Message send failed: missing contact and group_url")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "At least one of contact or group_url must be provided",
		})
	}

	if strings.TrimSpace(req.Platform) == "" {
		logger.Info("Message send failed: missing platform")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: platform",
		})
	}

	if strings.TrimSpace(req.Text) == "" && fileContent == "" && req.FileContent == "" {
		logger.Info("Message send failed: missing text and file")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Either text or file must be provided",
		})
	}

	if fileContent == "" && req.FileContent != "" {
		fileContent = req.FileContent
		fileExtension = req.FileExtension
		if fileExtension == "" {
			logger.Info("Message send failed: file_extension is required when file_content is provided")
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "file_extension is required when file_content is provided",
			})
		}
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

	var addresses []string
	var sendToType string

	if req.GroupURL != "" {
		for groupURL := range strings.SplitSeq(req.GroupURL, ",") {
			groupURL = strings.TrimSpace(groupURL)
			if groupURL != "" {
				if _, err := url.ParseRequestURI(groupURL); err != nil {
					logger.Info(fmt.Sprintf("Message send failed: invalid group_url '%s' - %v", groupURL, err))
					return c.JSON(http.StatusBadRequest, ErrorResponse{
						Error: fmt.Sprintf("Invalid group_url: %s", groupURL),
					})
				}
				addresses = append(addresses, groupURL)
			}
		}
		sendToType = "group_url"
	} else if req.Contact != "" {
		for contact := range strings.SplitSeq(req.Contact, ",") {
			contact = strings.TrimSpace(contact)
			if contact != "" {
				addresses = append(addresses, contact)
			}
		}
		sendToType = "contact"
	}

	if len(addresses) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "At least one valid contact or group_url must be provided",
		})
	}

	for _, address := range addresses {
		message := queuedMessage{
			DeviceID:      deviceID,
			PlatformName:  req.Platform,
			Text:          req.Text,
			Username:      matrixUsername,
			FileContent:   fileContent,
			FileExtension: fileExtension,
			ReplyID:       req.ReplyID,
		}

		if sendToType == "group_url" {
			message.GroupURL = address
		} else {
			message.Contact = address
		}

		if err := producer.Publish(exchangeName, routingKey, message, rabbitmq.DefaultPublishOptions()); err != nil {
			logger.Error(fmt.Sprintf("RabbitMQ message publish failed for %s %s: %v\n%s", sendToType, address, err, debug.Stack()))
			return echo.ErrInternalServerError
		}
	}

	logger.Info("Message(s) queued successfully")
	return c.JSON(http.StatusOK, SendMessageResponse{
		Message: "Message(s) queued successfully",
	})
}
