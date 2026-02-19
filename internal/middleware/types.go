package middleware

import "interface-api/internal/database"

type AuthMethod string

const (
	AuthMethodSession AuthMethod = "session"
	AuthMethodAPIKey  AuthMethod = "apikey"
)

type AuthMiddleware struct {
	db database.Service
}

func NewAuth(db database.Service) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}
