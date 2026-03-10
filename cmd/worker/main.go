package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "interface-api/pkg/config"
	"interface-api/pkg/logger"
	"interface-api/pkg/worker"
)

func main() {
	logger.Info("Starting standalone message worker service")

	w := worker.New()
	w.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down workers")
	w.Stop()
	logger.Info("All workers stopped")
}
