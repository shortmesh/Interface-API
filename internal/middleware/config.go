package middleware

import "os"

type TokenConfig struct {
	SessionPrefix string
	MatrixPrefix  string
}

func NewTokenConfig() *TokenConfig {
	sessionPrefix := os.Getenv("SESSION_TOKEN_PREFIX")
	if sessionPrefix == "" {
		sessionPrefix = "sk_"
	}

	matrixPrefix := os.Getenv("MATRIX_TOKEN_PREFIX")
	if matrixPrefix == "" {
		matrixPrefix = "mt_"
	}

	return &TokenConfig{
		SessionPrefix: sessionPrefix,
		MatrixPrefix:  matrixPrefix,
	}
}
