package authy

import (
	"interface-api/internal/api/v1/handlers/services/authy/otp"
	"interface-api/internal/database"
	"interface-api/internal/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group, db database.Service) {
	authMiddleware := middleware.NewBasicAuth(db, "authy")

	authyGroup := g.Group("/services/authy", authMiddleware.Authenticate())

	otpHandler := otp.NewHandler(db)
	otpGroup := authyGroup.Group("/otp")
	otpGroup.POST("/generate", otpHandler.Generate)
	otpGroup.POST("/verify", otpHandler.Verify)
}
