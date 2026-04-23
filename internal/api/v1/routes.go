package v1

import (
	"interface-api/internal/api/v1/handlers/adminsession"
	"interface-api/internal/api/v1/handlers/devices"
	"interface-api/internal/api/v1/handlers/tokens"
	"interface-api/internal/api/v1/handlers/webhooks"
	"interface-api/internal/database"
	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, db database.Service) {
	tokenHandler := tokens.NewTokenHandler(db)
	deviceWsHandler := devices.NewDeviceWebsocketHandler(db)
	webhookHandler := webhooks.NewWebhookHandler(db)
	adminSessionHandler := adminsession.NewAdminSessionHandler(db)
	bearerAuth := middleware.NewBearerAuth(db)
	basicAuth := middleware.NewBasicAuth()
	adminAuth := middleware.NewAdminAuth(db)

	// Token routes
	g.POST("/tokens", tokenHandler.Create, basicAuth.Authenticate())
	g.DELETE("/tokens/:id", tokenHandler.Delete, basicAuth.Authenticate())

	// Device routes
	g.POST("/devices", deviceWsHandler.Create, bearerAuth.Authenticate())
	g.GET("/devices", deviceWsHandler.List, bearerAuth.Authenticate())
	g.GET("/devices/qr-code", deviceWsHandler.QRCode, bearerAuth.AuthenticateWebSocket())
	g.DELETE("/devices", deviceWsHandler.Delete, bearerAuth.Authenticate())
	g.POST("/devices/:device_id/message", deviceWsHandler.SendMessage, bearerAuth.Authenticate())

	// Webhook routes
	g.POST("/webhooks", webhookHandler.Add, bearerAuth.Authenticate())
	g.GET("/webhooks", webhookHandler.List, bearerAuth.Authenticate())
	g.PUT("/webhooks/:id", webhookHandler.Update, bearerAuth.Authenticate())
	g.DELETE("/webhooks/:id", webhookHandler.Delete, bearerAuth.Authenticate())

	// Admin session routes
	adminGroup := g.Group("/admin")
	adminGroup.POST("/login", adminSessionHandler.Login)
	adminGroup.GET("/logout", adminSessionHandler.Logout, adminAuth.RequireAuth())
	adminGroup.GET("/tokens", adminSessionHandler.ListTokens, adminAuth.RequireAuth())
	adminGroup.POST("/matrix-token", adminSessionHandler.SetMatrixToken, adminAuth.RequireAuth())
	adminGroup.GET("/matrix-token-status", adminSessionHandler.CheckMatrixToken, adminAuth.RequireAuth())

	// Admin device endpoints (with matrix token injection)
	adminGroup.POST("/devices", deviceWsHandler.Create, adminAuth.RequireAuth(), adminAuth.InjectMatrixToken())
	adminGroup.GET("/devices", deviceWsHandler.List, adminAuth.RequireAuth(), adminAuth.InjectMatrixToken())
	adminGroup.DELETE("/devices", deviceWsHandler.Delete, adminAuth.RequireAuth(), adminAuth.InjectMatrixToken())
	adminGroup.POST("/devices/:device_id/message", deviceWsHandler.SendMessage, adminAuth.RequireAuth(), adminAuth.InjectMatrixToken())
	adminGroup.GET("/devices/qr-code", deviceWsHandler.QRCode, adminAuth.RequireAuth(), adminAuth.InjectMatrixToken())

	// Admin webhook endpoints (with matrix token injection)
	adminGroup.GET("/webhooks", webhookHandler.List, adminAuth.RequireAuth(), adminAuth.InjectMatrixToken())
	adminGroup.POST("/webhooks", webhookHandler.Add, adminAuth.RequireAuth(), adminAuth.InjectMatrixToken())
	adminGroup.PUT("/webhooks/:id", webhookHandler.Update, adminAuth.RequireAuth(), adminAuth.InjectMatrixToken())
	adminGroup.DELETE("/webhooks/:id", webhookHandler.Delete, adminAuth.RequireAuth(), adminAuth.InjectMatrixToken())
}
