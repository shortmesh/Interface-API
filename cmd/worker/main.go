package main

import (
	"os"
	"os/signal"
	"syscall"

	"interface-api/pkg/logger"
	"interface-api/pkg/worker"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env.default", ".env")

	logger.Log.Info("Starting standalone message worker service")

	w := worker.New()
	w.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Log.Info("Shutting down workers")
	w.Stop()
	logger.Log.Info("All workers stopped")
}
