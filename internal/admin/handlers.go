package admin

import (
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

	"interface-api/internal/api/v1/handlers/devices"
	"interface-api/internal/api/v1/handlers/tokens"
	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
)

//go:embed web/*
var webFS embed.FS

type AdminHandler struct {
	db            database.Service
	tokenHandler  *tokens.TokenHandler
	deviceHandler *devices.DeviceHandler
	sessions      map[string]*AdminSession
	sessionMutex  sync.RWMutex
}

type AdminSession struct {
	ExpiresAt   time.Time
	MatrixToken string
}

func NewAdminHandler(dbService database.Service) *AdminHandler {
	h := &AdminHandler{
		db:            dbService,
		tokenHandler:  tokens.NewTokenHandler(dbService),
		deviceHandler: devices.NewDeviceWebsocketHandler(dbService),
		sessions:      make(map[string]*AdminSession),
	}
	go h.cleanupExpiredSessions()
	return h
}

func (h *AdminHandler) cleanupExpiredSessions() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.sessionMutex.Lock()
		now := time.Now()
		for token, session := range h.sessions {
			if now.After(session.ExpiresAt) {
				delete(h.sessions, token)
			}
		}
		h.sessionMutex.Unlock()
	}
}

func (h *AdminHandler) generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *AdminHandler) storeSession(token string, expiration time.Time) {
	h.sessionMutex.Lock()
	defer h.sessionMutex.Unlock()
	h.sessions[token] = &AdminSession{
		ExpiresAt:   expiration,
		MatrixToken: "",
	}
}

func (h *AdminHandler) isSessionValid(token string) bool {
	h.sessionMutex.RLock()
	session, exists := h.sessions[token]
	h.sessionMutex.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		h.sessionMutex.Lock()
		delete(h.sessions, token)
		h.sessionMutex.Unlock()
		return false
	}
	return true
}

func (h *AdminHandler) clearSession(token string) {
	h.sessionMutex.Lock()
	defer h.sessionMutex.Unlock()
	delete(h.sessions, token)
}

func (h *AdminHandler) getMatrixToken(sessionToken string) string {
	h.sessionMutex.RLock()
	defer h.sessionMutex.RUnlock()
	if session, exists := h.sessions[sessionToken]; exists {
		return session.MatrixToken
	}
	return ""
}

func (h *AdminHandler) setMatrixToken(sessionToken, matrixToken string) error {
	h.sessionMutex.Lock()
	defer h.sessionMutex.Unlock()
	if session, exists := h.sessions[sessionToken]; exists {
		session.MatrixToken = matrixToken
		return nil
	}
	return fmt.Errorf("session not found")
}

func (h *AdminHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Admin credentials not configured"})
	}

	if subtle.ConstantTimeCompare([]byte(username), []byte(clientID)) != 1 ||
		subtle.ConstantTimeCompare([]byte(password), []byte(clientSecret)) != 1 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	sessionToken, err := h.generateSessionToken()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to generate session token: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Session creation failed"})
	}

	expiration := time.Now().Add(2 * time.Hour)
	h.storeSession(sessionToken, expiration)

	cookie := &http.Cookie{
		Name:     "shortmesh_admin_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("shortmesh_admin_token")
	if err == nil {
		h.clearSession(cookie.Value)
	}

	cookie = &http.Cookie{
		Name:     "shortmesh_admin_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/admin/login")
}

func (h *AdminHandler) listTokens(c echo.Context) error {
	var tokens []models.MatrixIdentity
	if err := h.db.DB().Find(&tokens).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to list tokens: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, tokens)
}

func (h *AdminHandler) deleteToken(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID is required"})
	}

	if err := h.db.DB().Delete(&models.MatrixIdentity{}, id).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to delete token: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	logger.Info(fmt.Sprintf("Token deleted: %s", id))
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *AdminHandler) setSessionMatrixToken(c echo.Context) error {
	cookie, err := c.Cookie("shortmesh_admin_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Session not found"})
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Token is required"})
	}

	_, err = models.FindMatrixIdentityByToken(h.db.DB(), req.Token)
	if err != nil {
		logger.Error(fmt.Sprintf("Invalid matrix token: %v", err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid matrix token"})
	}

	if err := h.setMatrixToken(cookie.Value, req.Token); err != nil {
		logger.Error(fmt.Sprintf("Failed to set matrix token: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set matrix token"})
	}

	logger.Info("Matrix token set for admin session")
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "message": "Matrix token attached to session"})
}

func (h *AdminHandler) checkSessionMatrixToken(c echo.Context) error {
	cookie, err := c.Cookie("shortmesh_admin_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Session not found"})
	}

	token := h.getMatrixToken(cookie.Value)
	if token == "" {
		return c.JSON(http.StatusOK, map[string]bool{"has_matrix_token": false})
	}
	return c.JSON(http.StatusOK, map[string]bool{"has_matrix_token": true})
}

func (h *AdminHandler) injectMatrixTokenFromSession(c echo.Context) error {
	cookie, err := c.Cookie("shortmesh_admin_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Session not found"})
	}

	token := h.getMatrixToken(cookie.Value)
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Matrix token not set for this session. Please set it first."})
	}

	matrixIdentity, err := models.FindMatrixIdentityByToken(h.db.DB(), token)
	if err != nil {
		logger.Error("Invalid matrix token in session")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid matrix token"})
	}

	c.Set("matrix_identity", matrixIdentity)
	c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return nil
}

func (h *AdminHandler) deviceWithMatrixToken(handler func(echo.Context) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := h.injectMatrixTokenFromSession(c); err != nil {
			return err
		}
		return handler(c)
	}
}

func (h *AdminHandler) checkAuth(c echo.Context) bool {
	cookie, err := c.Cookie("shortmesh_admin_token")
	if err != nil {
		return false
	}
	return h.isSessionValid(cookie.Value)
}

func (h *AdminHandler) requireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !h.checkAuth(c) {
			return c.Redirect(http.StatusFound, "/admin/login")
		}
		return next(c)
	}
}

func (h *AdminHandler) Index(c echo.Context) error {
	data, err := fs.ReadFile(webFS, "web/index.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "index.html not found"})
	}
	return c.HTML(http.StatusOK, string(data))
}

func (h *AdminHandler) TokensPage(c echo.Context) error {
	data, err := fs.ReadFile(webFS, "web/tokens.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "tokens.html not found"})
	}
	return c.HTML(http.StatusOK, string(data))
}

func (h *AdminHandler) DevicesPage(c echo.Context) error {
	data, err := fs.ReadFile(webFS, "web/devices.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "devices.html not found"})
	}
	return c.HTML(http.StatusOK, string(data))
}

func (h *AdminHandler) LoginPageStatic(c echo.Context) error {
	data, err := fs.ReadFile(webFS, "web/login.html")
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "login.html not found"})
	}
	return c.HTML(http.StatusOK, string(data))
}

func (h *AdminHandler) StaticFile(c echo.Context) error {
	file := c.Param("file")

	// Public files that don't require authentication
	if file == "styles.css" || file == "favicon.ico" || file == "whatsapp.png" || file == "signal.png" || file == "telegram.png" {
		data, err := fs.ReadFile(webFS, "web/"+file)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		if file == "styles.css" {
			return c.Blob(http.StatusOK, "text/css", data)
		}
		if file == "favicon.ico" {
			return c.Blob(http.StatusOK, "image/x-icon", data)
		}
		if file == "whatsapp.png" || file == "signal.png" || file == "telegram.png" {
			return c.Blob(http.StatusOK, "image/png", data)
		}
	}

	if !h.checkAuth(c) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
	}

	data, err := fs.ReadFile(webFS, "web/"+file)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	// Determine content type for JS files
	if file == "index.js" {
		return c.Blob(http.StatusOK, "application/javascript", data)
	}

	return c.Blob(http.StatusOK, "application/octet-stream", data)
}

func (h *AdminHandler) RegisterRoutes(e *echo.Echo) {
	admin := e.Group("/admin")
	admin.GET("/login", h.LoginPageStatic)
	admin.POST("/login", h.Login)
	admin.GET("/logout", h.Logout)

	admin.GET("", h.Index, h.requireAuth)
	admin.GET("/tokens", h.TokensPage, h.requireAuth)
	admin.GET("/devices", h.DevicesPage, h.requireAuth)
	admin.GET("/:file", h.StaticFile)

	adminAPI := e.Group("/api/v1/admin")
	adminAPI.GET("/tokens", h.listTokens, h.requireAuth)
	adminAPI.POST("/tokens", h.tokenHandler.Create, h.requireAuth)
	adminAPI.DELETE("/tokens/:id", h.deleteToken, h.requireAuth)

	// Matrix token session management
	adminAPI.POST("/matrix-token", h.setSessionMatrixToken, h.requireAuth)
	adminAPI.GET("/matrix-token-status", h.checkSessionMatrixToken, h.requireAuth)

	// Device endpoints (protected by session cookie and matrix token)
	adminAPI.POST("/devices", h.deviceWithMatrixToken(h.deviceHandler.Create), h.requireAuth)
	adminAPI.GET("/devices", h.deviceWithMatrixToken(h.deviceHandler.List), h.requireAuth)
	adminAPI.DELETE("/devices", h.deviceWithMatrixToken(h.deviceHandler.Delete), h.requireAuth)
	adminAPI.POST("/devices/:device_id/message", h.deviceWithMatrixToken(h.deviceHandler.SendMessage), h.requireAuth)
	adminAPI.GET("/devices/qr-code", h.deviceWithMatrixToken(h.deviceHandler.QRCode), h.requireAuth)
}
