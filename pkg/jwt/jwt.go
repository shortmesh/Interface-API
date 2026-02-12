package jwt

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"interface-api/internal/logger"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrMissingJWTKey    = errors.New("JWT_SECRET not configured")
	ErrInvalidKeyFormat = errors.New("JWT_SECRET must be base64 encoded")
	ErrInvalidKeySize   = errors.New("JWT_SECRET must be exactly 32 bytes when decoded")
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return
	}

	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		logger.Log.Errorf("Failed to load JWT_SECRET: %v: %v", ErrInvalidKeyFormat, err)
		return
	}

	if len(key) != 32 {
		logger.Log.Errorf("Failed to load JWT_SECRET: %v (got %d bytes)", ErrInvalidKeySize, len(key))
		return
	}

	jwtSecret = key
}

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	Token     string    `json:"token"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, sessionID uuid.UUID, sessionToken string, expiresAt time.Time) (string, error) {
	if jwtSecret == nil {
		return "", ErrMissingJWTKey
	}

	claims := &Claims{
		UserID:    userID,
		SessionID: sessionID,
		Token:     sessionToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	if jwtSecret == nil {
		return nil, ErrMissingJWTKey
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if time.Now().UTC().After(claims.ExpiresAt.Time) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}
