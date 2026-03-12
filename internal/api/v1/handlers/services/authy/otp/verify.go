package otp

import (
	"fmt"
	"net/http"
	"strings"

	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
	"interface-api/pkg/phoneutil"

	"github.com/labstack/echo/v4"
)

// Verify godoc
//
//	@Summary		Verify OTP
//	@Description	Verify an OTP code for the authenticated user
//	@Tags			services - authy
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Param			Authorization	header		string				true	"Basic Auth credentials"
//	@Param			X-Matrix-Token	header		string				true	"Matrix token (e.g., mt_xxxxx)"
//	@Param			request			body		VerifyOTPRequest	true	"OTP verification request"
//	@Success		200				{object}	VerifyOTPResponse	"OTP verified successfully"
//	@Failure		400				{object}	ErrorResponse		"Invalid request body or validation error"
//	@Failure		401				{object}	ErrorResponse		"Unauthorized"
//	@Failure		403				{object}	ErrorResponse		"Invalid or expired matrix token"
//	@Failure		429				{object}	ErrorResponse		"TooManyRequests"
//	@Failure		500				{object}	ErrorResponse		"Internal server error"
//	@Router			/api/v1/services/authy/otp/verify [post]
func (h *Handler) Verify(c echo.Context) error {
	userService, ok := c.Get("user_service").(*models.UserService)
	if !ok {
		logger.Error("User service not found in context")
		return echo.ErrUnauthorized
	}

	user, ok := c.Get("user").(*models.User)
	if !ok {
		logger.Error("User not found in context")
		return echo.ErrUnauthorized
	}

	var req VerifyOTPRequest
	if err := c.Bind(&req); err != nil {
		logger.Info(fmt.Sprintf("OTP verification failed: invalid request body - %v", err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Must be a JSON object.",
		})
	}

	if strings.TrimSpace(req.Identifier) == "" {
		logger.Info("OTP verification failed: missing identifier")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: identifier",
		})
	}

	if err := phoneutil.ValidateE164(req.Identifier); err != nil {
		logger.Info("OTP verification failed: identifier not in E.164 format")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "identifier must be in E.164 format (e.g., +1234567890)",
		})
	}

	if strings.TrimSpace(req.Platform) == "" {
		logger.Info("OTP verification failed: missing platform")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: platform",
		})
	}

	if strings.TrimSpace(req.Sender) == "" {
		logger.Info("OTP verification failed: missing sender")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: sender",
		})
	}

	if strings.TrimSpace(req.Code) == "" {
		logger.Info("OTP verification failed: missing code")
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing required field: code",
		})
	}

	err := models.VerifyOTP(
		h.db.DB(), userService.ID, req.Identifier, req.Platform, req.Sender, req.Code,
	)
	if err != nil {
		switch err {
		case models.ErrOTPNotFound:
			logger.Info("OTP verification failed: OTP not found")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid or expired OTP",
			})
		case models.ErrOTPExpired:
			logger.Info("OTP verification failed: OTP has expired")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid or expired OTP",
			})
		case models.ErrOTPInvalidCode:
			logger.Info("OTP verification failed: invalid OTP code")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid or expired OTP",
			})
		case models.ErrOTPTooManyAttempts:
			logger.Info("OTP verification failed: too many attempts")
			return c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Error: "Too many attempts, please request a new code",
			})
		case models.ErrOTPInvalidated:
			logger.Info("OTP verification failed: OTP was invalidated")
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid or expired OTP",
			})
		default:
			logger.Error(fmt.Sprintf("OTP verification error: %v", err))
			return echo.ErrInternalServerError
		}
	}

	logger.Info(fmt.Sprintf("OTP verified successfully for user %d (service: %d)", user.ID, userService.ID))
	return c.JSON(http.StatusOK, VerifyOTPResponse{
		Message: "OTP verified successfully",
	})
}
