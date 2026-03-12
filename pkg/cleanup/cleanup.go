package cleanup

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"interface-api/internal/database"
	"interface-api/internal/database/models"
	"interface-api/pkg/logger"
)

type CleanupWorker struct {
	ctx                 context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
	db                  database.Service
	otpInterval         time.Duration
	matrixTokenInterval time.Duration
}

func New() *CleanupWorker {
	otpIntervalMinutes := 15
	if interval := os.Getenv("OTP_CLEANUP_INTERVAL_MINUTES"); interval != "" {
		if n, err := strconv.Atoi(interval); err == nil && n > 0 {
			otpIntervalMinutes = n
		}
	}

	matrixTokenIntervalMinutes := 60
	if interval := os.Getenv("MATRIX_TOKEN_CLEANUP_INTERVAL_MINUTES"); interval != "" {
		if n, err := strconv.Atoi(interval); err == nil && n > 0 {
			matrixTokenIntervalMinutes = n
		}
	}

	db := database.New()
	ctx, cancel := context.WithCancel(context.Background())

	return &CleanupWorker{
		ctx:                 ctx,
		cancel:              cancel,
		db:                  db,
		otpInterval:         time.Duration(otpIntervalMinutes) * time.Minute,
		matrixTokenInterval: time.Duration(matrixTokenIntervalMinutes) * time.Minute,
	}
}

func IsEnabled() bool {
	return os.Getenv("CLEANUP_ENABLED") != "false"
}

func (cw *CleanupWorker) Start() {
	logger.Info(fmt.Sprintf("Starting cleanup worker - OTP interval: %v, Matrix token interval: %v", cw.otpInterval, cw.matrixTokenInterval))

	cw.wg.Add(2)
	go func() {
		defer cw.wg.Done()
		cw.runOTPCleanup()
	}()
	go func() {
		defer cw.wg.Done()
		cw.runMatrixTokenCleanup()
	}()
}

func (cw *CleanupWorker) Stop() {
	logger.Info("Stopping cleanup worker")
	cw.cancel()
	cw.wg.Wait()
	logger.Info("Cleanup worker stopped")
}

func (cw *CleanupWorker) runOTPCleanup() {
	ticker := time.NewTicker(cw.otpInterval)
	defer ticker.Stop()

	cw.cleanupOTPs()

	for {
		select {
		case <-cw.ctx.Done():
			return
		case <-ticker.C:
			cw.cleanupOTPs()
		}
	}
}

func (cw *CleanupWorker) runMatrixTokenCleanup() {
	ticker := time.NewTicker(cw.matrixTokenInterval)
	defer ticker.Stop()

	cw.cleanupMatrixTokens()

	for {
		select {
		case <-cw.ctx.Done():
			return
		case <-ticker.C:
			cw.cleanupMatrixTokens()
		}
	}
}

func (cw *CleanupWorker) cleanupOTPs() {
	result := cw.db.DB().Where("expires_at <= ?", time.Now().UTC()).Delete(&models.OTP{})
	if result.Error != nil {
		logger.Error(fmt.Sprintf("Failed to cleanup expired OTPs: %v", result.Error))
	} else if result.RowsAffected > 0 {
		logger.Info(fmt.Sprintf("Cleaned up %d expired OTP(s)", result.RowsAffected))
	}
}

func (cw *CleanupWorker) cleanupMatrixTokens() {
	result := cw.db.DB().Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now().UTC()).Delete(&models.MatrixIdentity{})
	if result.Error != nil {
		logger.Error(fmt.Sprintf("Failed to cleanup expired matrix tokens: %v", result.Error))
	} else if result.RowsAffected > 0 {
		logger.Info(fmt.Sprintf("Cleaned up %d expired matrix token(s)", result.RowsAffected))
	}
}
