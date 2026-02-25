package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"interface-api/internal/server"
	"interface-api/pkg/logger"
	"interface-api/pkg/worker"

	"github.com/joho/godotenv"
)

//	@title			Shortmesh - Interface API
//	@version		1.0
//	@description	API for ShortMesh Interface service

//	@host		localhost:8080
//	@schemes	http

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter your token in the format: Bearer {token}

func gracefulShutdown(apiServer *http.Server, w *worker.Worker, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger.Log.Info("Shutting down gracefully, press Ctrl+C again to force")
	stop()

	if w != nil {
		w.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Log.Errorf("Server shutdown error: %v", err)
	}

	logger.Log.Info("Server exiting")

	done <- true
}

func main() {
	godotenv.Load(".env.default", ".env")

	var w *worker.Worker
	if worker.IsEnabled() {
		w = worker.New()
		w.Start()
	} else {
		logger.Log.Info("Worker disabled via WORKER_ENABLED=false")
	}

	srv := server.NewServer()

	logger.Log.Infof("Starting server on %s", srv.Addr)

	done := make(chan bool, 1)

	go gracefulShutdown(srv, w, done)

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Log.Fatalf("Server startup failed: %v", err)
	}

	<-done
	logger.Log.Info("Graceful shutdown complete")
}
