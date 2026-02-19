package apikeys

import "interface-api/internal/database"

// CreateAPIKeyRequest represents the request body for creating an API key
type CreateAPIKeyRequest struct {
	Name      string  `json:"name" example:"Production API Key" validate:"required,min=1,max=255"`
	ExpiresAt *string `json:"expires_at,omitempty" example:"2026-12-31T23:59:59Z"`
}

// APIKeyResponse represents a newly created API key with the token
type APIKeyResponse struct {
	Message string      `json:"message,omitempty" example:"API key created successfully"`
	Key     string      `json:"key,omitempty" example:"ak_123456789abcdef123456789abcdef123456789"`
	APIKey  *APIKeyInfo `json:"data,omitempty"`
}

// APIKeyInfo represents API key metadata
type APIKeyInfo struct {
	KeyID      string  `json:"key_id" example:"a1b2c3d4e5f6g7h8"`
	Name       string  `json:"name" example:"Production API Key"`
	ExpiresAt  *string `json:"expires_at,omitempty" example:"2026-12-31T23:59:59Z"`
	CreatedAt  string  `json:"created_at" example:"2026-02-19T20:00:00Z"`
	LastUsedAt *string `json:"last_used_at,omitempty" example:"2026-02-19T20:30:00Z"`
}

// ListAPIKeysResponse represents the response for listing API keys
type ListAPIKeysResponse struct {
	APIKeys []APIKeyInfo `json:"data"`
}

// DeleteAPIKeyRequest represents the request body for deleting an API key
type DeleteAPIKeyRequest struct {
	KeyID string `json:"key_id" example:"a1b2c3d4e5f6g7h8" validate:"required"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"API key deleted successfully"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"message"`
}

type APIKeyHandler struct {
	db database.Service
}

func NewAPIKeyHandler(db database.Service) *APIKeyHandler {
	return &APIKeyHandler{db: db}
}
