package otp

import (
	"os"

	"interface-api/internal/database"
)

type GenerateOTPRequest struct {
	Identifier string `json:"identifier" example:"237123456780" validate:"required"`
	Platform   string `json:"platform" example:"wa" validate:"required"`
	Sender     string `json:"sender" example:"237123456789" validate:"required"`
}

type GenerateOTPResponse struct {
	Message   string `json:"message" example:"OTP sent successfully"`
	ExpiresAt string `json:"expires_at" example:"2026-02-19T20:30:00Z"`
}

type VerifyOTPRequest struct {
	Identifier string `json:"identifier" example:"237123456780" validate:"required"`
	Platform   string `json:"platform" example:"wa" validate:"required"`
	Sender     string `json:"sender" example:"237123456789" validate:"required"`
	Code       string `json:"code" example:"123456" validate:"required"`
}

type VerifyOTPResponse struct {
	Message string `json:"message" example:"OTP verified successfully"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"message"`
}

type Handler struct {
	db        database.Service
	rabbitURL *string
}

func NewHandler(db database.Service) *Handler {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	return &Handler{
		db:        db,
		rabbitURL: &rabbitURL,
	}
}
