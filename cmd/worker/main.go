// cmd/worker/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ega-telkom/fundivest-backend/internal/config"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
    	log.Printf("Failed to load .env: %v", err)
	}

	cfg := config.Load()
	logger := config.NewLogger(cfg)
	defer func() { _ = logger.Sync() }()

	logger.Info("Starting worker")

	infra := SetupInfrastructure(cfg, logger)
	defer infra.Close()

	worker := SetupWorker(cfg, infra, logger)

	go func() {
		logger.Info("Worker started, waiting for jobs...")
		if err := worker.Start(); err != nil {
			logger.Fatal("Worker failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down worker...")
	worker.Shutdown()
}
