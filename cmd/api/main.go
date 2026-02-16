package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"interface-api/internal/logger"
	"interface-api/internal/server"
)

// @title Interface API
// @version 1.0
// @description API for ShortMesh Interface service

// @host localhost:8080
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your token in the format: Bearer {token}

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Log.Info("shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Log.Errorf("Server forced to shutdown with error: %v", err)
	}

	logger.Log.Info("Server exiting")

	done <- true
}

func main() {
	logger.Init()

	srv := server.NewServer()

	logger.Log.Infof("Starting server on %s", srv.Addr)

	done := make(chan bool, 1)

	go gracefulShutdown(srv, done)

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Log.Fatalf("Server failed to start: %v", err)
	}

	<-done
	logger.Log.Info("Graceful shutdown complete.")
}
