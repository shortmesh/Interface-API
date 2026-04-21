package webhooks

import "interface-api/internal/database"

type WebhookHandler struct {
	db database.Service
}

func NewWebhookHandler(db database.Service) *WebhookHandler {
	return &WebhookHandler{db: db}
}

type AddWebhookRequest struct {
	URL string `json:"url"`
}

type UpdateWebhookRequest struct {
	URL    *string `json:"url,omitempty"`
	Active *bool   `json:"active,omitempty"`
}

type WebhookResponse struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
