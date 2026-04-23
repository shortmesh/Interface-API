package adminsession

import (
	"interface-api/internal/database"
)

type AdminSessionHandler struct {
	db database.Service
}

func NewAdminSessionHandler(db database.Service) *AdminSessionHandler {
	return &AdminSessionHandler{db: db}
}

type LoginRequest struct {
	Username string `form:"username" json:"username" example:"admin"`
	Password string `form:"password" json:"password" example:"password123"`
}

type LoginResponse struct {
	Status string `json:"status" example:"ok"`
}

type LogoutResponse struct {
	Message string `json:"message" example:"Logged out successfully"`
}

type SetMatrixTokenRequest struct {
	Token string `json:"token" example:"mt_xxxxx"`
}

type SetMatrixTokenResponse struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message" example:"Matrix token attached to session"`
}

type MatrixTokenStatusResponse struct {
	HasMatrixToken bool `json:"has_matrix_token" example:"true"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid credentials"`
}
