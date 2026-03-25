// cmd/worker/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ega-telkom/fundivest-backend/internal/config"
	"github.com/Ega-telkom/fundivest-backend/internal/pubsub"
	"github.com/valkey-io/valkey-go"

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
	valkeyClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.ValkeyAddr()},
		Password: cfg.ValkeyPassword,
	})
	if err != nil {
		logger.Fatal("Failed to connect to Valkey", zap.Error(err))
	}
	logger.Info("Connected to Valkey")
	
	pubsub := pubsub.NewRedisPubSub(valkeyClient)
	
	defer infra.Close()

	worker := SetupWorker(cfg, infra, pubsub, logger)

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
