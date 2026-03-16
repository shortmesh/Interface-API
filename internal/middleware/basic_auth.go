package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

type BasicAuthMiddleware struct {
	clientID     string
	clientSecret string
}

func NewBasicAuth() *BasicAuthMiddleware {
	return &BasicAuthMiddleware{
		clientID:     os.Getenv("CLIENT_ID"),
		clientSecret: os.Getenv("CLIENT_SECRET"),
	}
}

func (m *BasicAuthMiddleware) Authenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "basic" {
				logger.Error(fmt.Sprintf("Expected Basic auth, got: %s", authHeader))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}

			credParts := strings.SplitN(string(decoded), ":", 2)
			if len(credParts) != 2 {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}

			clientID := credParts[0]
			clientSecret := credParts[1]

			if subtle.ConstantTimeCompare([]byte(clientID), []byte(m.clientID)) != 1 ||
				subtle.ConstantTimeCompare([]byte(clientSecret), []byte(m.clientSecret)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}

			return next(c)
		}
	}
}
