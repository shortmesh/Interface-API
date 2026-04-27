package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"interface-api/internal/database"
	"interface-api/internal/server"
	"interface-api/pkg/cleanup"
	"interface-api/pkg/config"
	"interface-api/pkg/logger"
	"interface-api/pkg/webhookworker"
	"interface-api/pkg/worker"
)

//	@title			Shortmesh - Interface API
//	@version		1.0
//	@description	API for ShortMesh Interface service

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Bearer token authentication (use Bearer mt_xxxxx")

//	@securityDefinitions.basic	BasicAuth

//	@securityDefinitions.apikey	CookieAuth
//	@in							cookie
//	@name						shortmesh_admin_token
//	@description				Admin session cookie authentication

func gracefulShutdown(apiServer *http.Server, w *worker.Worker, cw *cleanup.CleanupWorker, ww *webhookworker.WebhookWorker, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Info("Shutting down gracefully, press Ctrl+C again to force")
	stop()

	if w != nil {
		w.Stop()
	}

	if cw != nil {
		cw.Stop()
	}

	if ww != nil {
		ww.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server shutdown error: %v", err))
	}

	logger.Info("Server exiting")

	done <- true
}

func main() {
	db := database.New()

	if err := database.InitializeSuperAdminCredentials(db.DB()); err != nil {
		logger.Error(fmt.Sprintf("Credential initialization failed: %v", err))
		os.Exit(1)
	}

	var w *worker.Worker
	if worker.IsEnabled() {
		w = worker.New()
		w.Start()
	} else {
		logger.Info("Worker disabled via WORKER_ENABLED=false")
	}

	var cw *cleanup.CleanupWorker
	if cleanup.IsEnabled() {
		cw = cleanup.New()
		cw.Start()
	} else {
		logger.Info("Cleanup worker disabled via CLEANUP_ENABLED=false")
	}

	var ww *webhookworker.WebhookWorker
	if webhookworker.IsEnabled() {
		ww = webhookworker.New(db)
		ww.Start()
	} else {
		logger.Info("Webhook worker disabled via WEBHOOK_WORKER_ENABLED=false")
	}

	srv := server.NewServer()

	done := make(chan bool, 1)

	go gracefulShutdown(srv, w, cw, ww, done)

	var err error
	if config.RequiresHTTPS() {
		certFile := os.Getenv("TLS_CERT_FILE")
		keyFile := os.Getenv("TLS_KEY_FILE")

		if certFile == "" || keyFile == "" {
			logger.Error("TLS_CERT_FILE and TLS_KEY_FILE must be set when running in production mode without ALLOW_INSECURE_SERVER")
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("Starting HTTPS server on %s", srv.Addr))
		err = srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		protocol := "HTTP"
		if config.IsProd() {
			protocol = "HTTP (insecure mode for reverse proxy)"
		}
		logger.Info(fmt.Sprintf("Starting %s server on %s", protocol, srv.Addr))
		err = srv.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		logger.Error(fmt.Sprintf("Server startup failed: %v", err))
		os.Exit(1)
	}

	<-done
	logger.Info("Graceful shutdown complete")
}
