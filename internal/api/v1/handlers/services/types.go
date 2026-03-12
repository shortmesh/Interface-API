package services

import "interface-api/internal/database"

type ServiceHandler struct {
	db database.Service
}

func NewServiceHandler(db database.Service) *ServiceHandler {
	return &ServiceHandler{db: db}
}

type ServiceInfo struct {
	Name        string `json:"name" example:"authy"`
	DisplayName string `json:"display_name" example:"Authy OTP Service"`
	Description string `json:"description" example:"One-time password generation and verification service"`
}

type ListServicesResponse struct {
	Services []ServiceInfo `json:"services"`
}

type UserServiceInfo struct {
	Name         string  `json:"name" example:"authy"`
	DisplayName  string  `json:"display_name" example:"Authy OTP Service"`
	Description  string  `json:"description" example:"One-time password generation and verification service"`
	IsEnabled    bool    `json:"is_enabled" example:"true"`
	IsExpired    bool    `json:"is_expired" example:"false"`
	ClientID     string  `json:"client_id" example:"your_client_id"`
	ClientSecret string  `json:"client_secret" example:"your_client_secret"`
	CreatedAt    string  `json:"created_at" example:"2026-03-12T10:00:00Z"`
	ExpiresAt    *string `json:"expires_at,omitempty" example:"2027-03-12T10:00:00Z"`
}

type ListUserServicesResponse struct {
	Services []UserServiceInfo `json:"services"`
}

type ServiceStatusResponse struct {
	Name         string `json:"name" example:"authy"`
	DisplayName  string `json:"display_name" example:"Authy OTP Service"`
	IsSubscribed bool   `json:"is_subscribed" example:"true"`
	IsEnabled    bool   `json:"is_enabled" example:"true"`
	IsExpired    bool   `json:"is_expired" example:"false"`
}

type SubscribeResponse struct {
	Message      string `json:"message" example:"Successfully subscribed to service"`
	ServiceName  string `json:"service_name" example:"authy"`
	ClientID     string `json:"client_id" example:"your_client_id"`
	ClientSecret string `json:"client_secret" example:"your_client_secret"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}
