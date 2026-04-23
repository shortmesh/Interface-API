package credentials

import (
	"interface-api/internal/database"
	"interface-api/internal/database/models"
)

type CredentialHandler struct {
	db database.Service
}

func NewCredentialHandler(db database.Service) *CredentialHandler {
	return &CredentialHandler{db: db}
}

type CreateRequest struct {
	ClientID    string `json:"client_id" validate:"required"`
	Description string `json:"description"`
}

type CreateResponse struct {
	Message      string              `json:"message"`
	Credential   *CredentialResponse `json:"credential"`
	ClientSecret string              `json:"client_secret"`
}

type ListResponse struct {
	Credentials []CredentialResponse `json:"credentials"`
}

type CredentialResponse struct {
	ClientID    string                `json:"client_id"`
	Role        models.CredentialRole `json:"role"`
	Scopes      models.Scopes         `json:"scopes"`
	Description string                `json:"description"`
	Active      bool                  `json:"active"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}

type UpdateRequest struct {
	RegenerateSecret *bool   `json:"regenerate_secret,omitempty"`
	Deactivate       *bool   `json:"deactivate,omitempty"`
	Description      *string `json:"description,omitempty"`
}

type UpdateResponse struct {
	Message      string              `json:"message"`
	Credential   *CredentialResponse `json:"credential"`
	ClientSecret *string             `json:"client_secret,omitempty"`
}

type DeleteResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
