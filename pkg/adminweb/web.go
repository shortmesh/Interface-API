package adminweb

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"interface-api/internal/database"
	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

type WebHandler struct {
	db database.Service
}

func NewWebHandler(db database.Service) *WebHandler {
	return &WebHandler{
		db: db,
	}
}

// ServeIndex serves the React app's index.html for all routes
func (h *WebHandler) ServeIndex(c echo.Context) error {
	data, err := fs.ReadFile(WebFS, "web/dist/index.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "index.html not found"})
	}
	return c.HTMLBlob(http.StatusOK, data)
}

// ServeStatic serves static assets from the React build
func (h *WebHandler) ServeStatic(c echo.Context) error {
	path := c.Request().URL.Path
	// Remove /admin/ prefix
	path = strings.TrimPrefix(path, "/admin/")

	// Security check: prevent directory traversal
	if strings.Contains(path, "..") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid path"})
	}

	// Try to read the file from dist
	filePath := filepath.Join("web/dist", path)
	data, err := fs.ReadFile(WebFS, filePath)
	if err != nil {
		// If file not found, serve index.html for React Router
		return h.ServeIndex(c)
	}

	// Determine content type based on file extension
	ext := filepath.Ext(path)
	contentType := "application/octet-stream"
	switch ext {
	case ".html":
		contentType = "text/html"
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	case ".json":
		contentType = "application/json"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".svg":
		contentType = "image/svg+xml"
	case ".ico":
		contentType = "image/x-icon"
	case ".woff":
		contentType = "font/woff"
	case ".woff2":
		contentType = "font/woff2"
	case ".ttf":
		contentType = "font/ttf"
	}

	return c.Blob(http.StatusOK, contentType, data)
}

// ProxyTokenCreate proxies token creation to /api/v1/tokens with Basic Auth
func (h *WebHandler) ProxyTokenCreate(c echo.Context) error {
	// Read the request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to read request body"})
	}

	// Create new request to internal API
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/tokens", bytes.NewReader(body))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create request"})
	}

	// Add Basic Auth from environment variables
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "CLIENT_ID or CLIENT_SECRET not configured"})
	}

	credentials := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to execute request"})
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read response"})
	}

	// Forward the response
	return c.JSONBlob(resp.StatusCode, respBody)
}

// ProxyTokenDelete proxies token deletion to /api/v1/tokens/:id with Basic Auth
func (h *WebHandler) ProxyTokenDelete(c echo.Context) error {
	tokenID := c.Param("id")

	// Create new request to internal API
	req, err := http.NewRequest("DELETE", "http://localhost:8080/api/v1/tokens/"+tokenID, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create request"})
	}

	// Add Basic Auth from environment variables
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "CLIENT_ID or CLIENT_SECRET not configured"})
	}

	credentials := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+credentials)

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to execute request"})
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read response"})
	}

	// Forward the response
	return c.JSONBlob(resp.StatusCode, respBody)
}

func (h *WebHandler) RegisterRoutes(e *echo.Echo) {
	adminAuth := middleware.NewAdminAuth(h.db)

	// IMPORTANT: Order matters! Register more specific routes FIRST

	// 1. Public static assets (no auth)
	e.GET("/admin/assets/*", h.ServeStatic)

	// 2. Public login page (no auth) - must be before catch-all
	e.GET("/admin/login", h.ServeIndex)
	e.GET("/admin/login/*", h.ServeIndex)

	// 3. API proxy endpoints (require admin auth) - BEFORE React routes
	e.POST("/admin/api/tokens", h.ProxyTokenCreate, adminAuth.RequireAuth())
	e.DELETE("/admin/api/tokens/:id", h.ProxyTokenDelete, adminAuth.RequireAuth())

	// 4. Protected specific routes (require auth)
	e.GET("/admin", h.ServeIndex, adminAuth.RequireAuth())
	e.GET("/admin/", h.ServeIndex, adminAuth.RequireAuth())
	e.GET("/admin/tokens", h.ServeIndex, adminAuth.RequireAuth())
	e.GET("/admin/tokens/*", h.ServeIndex, adminAuth.RequireAuth())
	e.GET("/admin/devices", h.ServeIndex, adminAuth.RequireAuth())
	e.GET("/admin/devices/*", h.ServeIndex, adminAuth.RequireAuth())
	e.GET("/admin/webhooks", h.ServeIndex, adminAuth.RequireAuth())
	e.GET("/admin/webhooks/*", h.ServeIndex, adminAuth.RequireAuth())
}
