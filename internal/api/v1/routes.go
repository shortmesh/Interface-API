package v1

import (
	"interface-api/internal/api/v1/handlers/adminsession"
	"interface-api/internal/api/v1/handlers/credentials"
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
	credentialHandler := credentials.NewCredentialHandler(db)

	bearerAuth := middleware.NewBearerAuth(db)
	credentialAuth := middleware.NewCredentialAuth(db)
	adminAuth := middleware.NewAdminAuth(db)

	// Credentials
	g.POST(
		"/credentials",
		credentialHandler.Create,
		credentialAuth.Authenticate(),
		credentialAuth.RequireScope("credentials:write:*"),
	)
	g.GET(
		"/credentials",
		credentialHandler.List,
		credentialAuth.Authenticate(),
		credentialAuth.RequireScope("credentials:read:*"),
	)
	g.PUT(
		"/credentials/:client_id",
		credentialHandler.Update,
		credentialAuth.Authenticate(),
		credentialAuth.RequireScope("credentials:write:*"),
	)
	g.DELETE(
		"/credentials/:client_id",
		credentialHandler.Delete,
		credentialAuth.Authenticate(),
		credentialAuth.RequireScope("credentials:write:*"),
	)

	// Tokens
	g.POST(
		"/tokens",
		tokenHandler.Create,
		credentialAuth.Authenticate(),
		credentialAuth.RequireScope("tokens:write:create"),
	)
	g.GET(
		"/tokens",
		tokenHandler.List,
		credentialAuth.Authenticate(),
		credentialAuth.RequireScope("tokens:read:*"),
	)
	g.DELETE(
		"/tokens/:id",
		tokenHandler.Delete,
		credentialAuth.Authenticate(),
		credentialAuth.RequireScope("tokens:write:delete"),
	)

	// Devices
	g.POST("/devices", deviceWsHandler.Create, bearerAuth.Authenticate())
	g.GET("/devices", deviceWsHandler.List, bearerAuth.Authenticate())
	g.GET("/devices/qr-code", deviceWsHandler.QRCode, bearerAuth.AuthenticateWebSocket())
	g.DELETE("/devices", deviceWsHandler.Delete, bearerAuth.Authenticate())
	g.POST(
		"/devices/:device_id/message",
		deviceWsHandler.SendMessage,
		bearerAuth.Authenticate(),
	)

	// Webhooks
	g.POST("/webhooks", webhookHandler.Add, bearerAuth.Authenticate())
	g.GET("/webhooks", webhookHandler.List, bearerAuth.Authenticate())
	g.PUT("/webhooks/:id", webhookHandler.Update, bearerAuth.Authenticate())
	g.DELETE("/webhooks/:id", webhookHandler.Delete, bearerAuth.Authenticate())

	// Admin routes
	adminGroup := g.Group("/admin")
	adminGroup.POST("/login", adminSessionHandler.Login)
	adminGroup.GET("/logout", adminSessionHandler.Logout, adminAuth.RequireAuth())
	adminGroup.POST(
		"/matrix-token",
		adminSessionHandler.SetMatrixToken,
		adminAuth.RequireAuth(),
	)
	adminGroup.GET(
		"/matrix-token-status",
		adminSessionHandler.CheckMatrixToken,
		adminAuth.RequireAuth(),
	)

	adminGroup.POST(
		"/credentials",
		credentialHandler.Create,
		adminAuth.RequireAuth(),
		adminAuth.InjectCredential(),
		credentialAuth.RequireScope("credentials:write:*"),
	)
	adminGroup.GET(
		"/credentials",
		credentialHandler.List,
		adminAuth.RequireAuth(),
		adminAuth.InjectCredential(),
		credentialAuth.RequireScope("credentials:read:*"),
	)
	adminGroup.PUT(
		"/credentials/:client_id",
		credentialHandler.Update,
		adminAuth.RequireAuth(),
		adminAuth.InjectCredential(),
		credentialAuth.RequireScope("credentials:write:*"),
	)
	adminGroup.DELETE(
		"/credentials/:client_id",
		credentialHandler.Delete,
		adminAuth.RequireAuth(),
		adminAuth.InjectCredential(),
		credentialAuth.RequireScope("credentials:write:*"),
	)

	adminGroup.POST(
		"/tokens",
		tokenHandler.Create,
		adminAuth.RequireAuth(),
		adminAuth.InjectCredential(),
		credentialAuth.RequireScope("tokens:write:create"),
	)
	adminGroup.GET(
		"/tokens",
		tokenHandler.List,
		adminAuth.RequireAuth(),
		adminAuth.InjectCredential(),
		credentialAuth.RequireScope("tokens:read:*"),
	)
	adminGroup.DELETE(
		"/tokens/:id",
		tokenHandler.Delete,
		adminAuth.RequireAuth(),
		adminAuth.InjectCredential(),
		credentialAuth.RequireScope("tokens:write:delete"),
	)

	adminGroup.POST(
		"/devices",
		deviceWsHandler.Create,
		adminAuth.RequireAuth(),
		adminAuth.InjectMatrixToken(),
	)
	adminGroup.GET(
		"/devices",
		deviceWsHandler.List,
		adminAuth.RequireAuth(),
		adminAuth.InjectMatrixToken(),
	)
	adminGroup.GET(
		"/devices/qr-code",
		deviceWsHandler.QRCode,
		adminAuth.RequireAuth(),
		adminAuth.InjectMatrixToken(),
	)
	adminGroup.DELETE(
		"/devices",
		deviceWsHandler.Delete,
		adminAuth.RequireAuth(),
		adminAuth.InjectMatrixToken(),
	)
	adminGroup.POST(
		"/devices/:device_id/message",
		deviceWsHandler.SendMessage,
		adminAuth.RequireAuth(),
		adminAuth.InjectMatrixToken(),
	)

	adminGroup.GET(
		"/webhooks",
		webhookHandler.List,
		adminAuth.RequireAuth(),
		adminAuth.InjectMatrixToken(),
	)
	adminGroup.POST(
		"/webhooks",
		webhookHandler.Add,
		adminAuth.RequireAuth(),
		adminAuth.InjectMatrixToken(),
	)
	adminGroup.PUT(
		"/webhooks/:id",
		webhookHandler.Update,
		adminAuth.RequireAuth(),
		adminAuth.InjectMatrixToken(),
	)
	adminGroup.DELETE(
		"/webhooks/:id",
		webhookHandler.Delete,
		adminAuth.RequireAuth(),
		adminAuth.InjectMatrixToken(),
	)
}
