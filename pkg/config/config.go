package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"interface-api/pkg/logger"

	"github.com/joho/godotenv"
)

type Mode string

const (
	ModeDevelopment Mode = "development"
	ModeProduction  Mode = "production"
)

type Config struct {
	Mode                  Mode
	RequireHTTPS          bool
	RequireHTTPSExternal  bool
	AllowInsecureExternal bool
	AllowInsecureServer   bool
}

var appConfig *Config

func init() {
	if os.Getenv("APP_MODE") != "production" {
		godotenv.Load("default.env", ".env")
	}
	load()
}

func load() {
	mode := Mode(strings.ToLower(os.Getenv("APP_MODE")))
	if mode == "" {
		mode = ModeDevelopment
	}

	if mode != ModeDevelopment && mode != ModeProduction {
		logger.Warn(fmt.Sprintf("Invalid APP_MODE '%s', defaulting to 'development'", mode))
		mode = ModeDevelopment
	}

	cfg := &Config{
		Mode: mode,
	}

	if mode == ModeProduction {
		cfg.RequireHTTPS = true
		cfg.RequireHTTPSExternal = true

		if getBoolEnv("ALLOW_INSECURE_SERVER", false) {
			cfg.AllowInsecureServer = true
			logger.Warn("SECURITY WARNING: ALLOW_INSECURE_SERVER enabled in production - use only behind reverse proxy with TLS termination")
		}

		if getBoolEnv("ALLOW_INSECURE_EXTERNAL", false) {
			cfg.AllowInsecureExternal = true
			logger.Warn("SECURITY WARNING: ALLOW_INSECURE_EXTERNAL enabled in production - external services will accept non-HTTPS connections")
		}
	} else {
		cfg.RequireHTTPS = false
		cfg.RequireHTTPSExternal = false
		cfg.AllowInsecureServer = true
		cfg.AllowInsecureExternal = true
	}

	appConfig = cfg
	logger.Info(fmt.Sprintf("Application mode: %s", mode))
}

func Get() *Config {
	return appConfig
}

func IsProd() bool {
	return Get().Mode == ModeProduction
}

func IsDev() bool {
	return Get().Mode == ModeDevelopment
}

func RequiresHTTPS() bool {
	cfg := Get()
	return cfg.RequireHTTPS && !cfg.AllowInsecureServer
}

func RequiresHTTPSExternal() bool {
	cfg := Get()
	return cfg.RequireHTTPSExternal && !cfg.AllowInsecureExternal
}

func ValidateExternalURL(url, serviceName string) error {
	if !RequiresHTTPSExternal() {
		return nil
	}

	urlLower := strings.ToLower(url)
	if !strings.HasPrefix(urlLower, "https://") &&
		!strings.HasPrefix(urlLower, "wss://") &&
		!strings.HasPrefix(urlLower, "amqps://") {
		return fmt.Errorf(
			"production mode requires HTTPS/WSS/AMQPS for '%s' (got: %s). Set ALLOW_INSECURE_EXTERNAL=true to override",
			serviceName, url,
		)
	}

	return nil
}

func getBoolEnv(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}
