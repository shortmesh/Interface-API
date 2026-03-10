package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"interface-api/internal/server"
	"interface-api/pkg/config"
	"interface-api/pkg/logger"
	"interface-api/pkg/worker"
)

//	@title			Shortmesh - Interface API
//	@version		1.0
//	@description	API for ShortMesh Interface service

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization

func gracefulShutdown(apiServer *http.Server, w *worker.Worker, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Info("Shutting down gracefully, press Ctrl+C again to force")
	stop()

	if w != nil {
		w.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server shutdown error: %v", err))
	}

	logger.Info("Server exiting")

	done <- true
}

func main() {
	var w *worker.Worker
	if worker.IsEnabled() {
		w = worker.New()
		w.Start()
	} else {
		logger.Info("Worker disabled via WORKER_ENABLED=false")
	}

	srv := server.NewServer()

	done := make(chan bool, 1)

	go gracefulShutdown(srv, w, done)

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
