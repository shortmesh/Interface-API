package tokens

import "interface-api/internal/database"

type TokenHandler struct {
	db database.Service
}

func NewTokenHandler(db database.Service) *TokenHandler {
	return &TokenHandler{db: db}
}

type CreateRequest struct {
	ExpiresAt *string `json:"expires_at,omitempty" example:"2026-12-31T23:59:59Z"`
}

type CreateResponse struct {
	Message string `json:"message"`
	Token   string `json:"token" example:"mt_xxxxx"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
