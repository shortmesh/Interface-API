package adminweb

import (
	"io/fs"
	"net/http"

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

func (h *WebHandler) Index(c echo.Context) error {
	data, err := fs.ReadFile(WebFS, "web/index.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "index.html not found"})
	}
	return c.HTML(http.StatusOK, string(data))
}

func (h *WebHandler) TokensPage(c echo.Context) error {
	data, err := fs.ReadFile(WebFS, "web/tokens.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "tokens.html not found"})
	}
	return c.HTML(http.StatusOK, string(data))
}

func (h *WebHandler) DevicesPage(c echo.Context) error {
	data, err := fs.ReadFile(WebFS, "web/devices.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "devices.html not found"})
	}
	return c.HTML(http.StatusOK, string(data))
}

func (h *WebHandler) WebhooksPage(c echo.Context) error {
	data, err := fs.ReadFile(WebFS, "web/webhooks.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "webhooks.html not found"})
	}
	return c.HTML(http.StatusOK, string(data))
}

func (h *WebHandler) LoginPage(c echo.Context) error {
	data, err := fs.ReadFile(WebFS, "web/login.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "login.html not found"})
	}
	return c.HTML(http.StatusOK, string(data))
}

func (h *WebHandler) StaticFile(c echo.Context) error {
	file := c.Param("file")

	publicFiles := map[string]string{
		"styles.css":   "text/css",
		"favicon.ico":  "image/x-icon",
		"whatsapp.png": "image/png",
		"signal.png":   "image/png",
		"telegram.png": "image/png",
	}

	if contentType, isPublic := publicFiles[file]; isPublic {
		data, err := fs.ReadFile(WebFS, "web/"+file)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.Blob(http.StatusOK, contentType, data)
	}

	cookie, err := c.Cookie("shortmesh_admin_token")
	if err != nil || !middleware.IsSessionValid(cookie.Value) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
	}

	data, err := fs.ReadFile(WebFS, "web/"+file)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	if file == "index.js" {
		return c.Blob(http.StatusOK, "application/javascript", data)
	}

	return c.Blob(http.StatusOK, "application/octet-stream", data)
}

func (h *WebHandler) RegisterRoutes(e *echo.Echo) {
	adminAuth := middleware.NewAdminAuth(h.db)

	admin := e.Group("/admin")
	admin.GET("/login", h.LoginPage)
	admin.GET("", h.Index, adminAuth.RequireAuth())
	admin.GET("/tokens", h.TokensPage, adminAuth.RequireAuth())
	admin.GET("/devices", h.DevicesPage, adminAuth.RequireAuth())
	admin.GET("/webhooks", h.WebhooksPage, adminAuth.RequireAuth())
	admin.GET("/:file", h.StaticFile)
}
