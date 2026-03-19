// cmd/api/main.go
package main

import (
	"os"
	"time"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/Ega-telkom/fundivest-backend/internal/config"
)

// @title           fundivest REST API
// @version         1.0
// @description     Manajemen Sesi & Sertifikat untuk fundivest
//
// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	// Load .env only in non-production
    // In production, vars come from docker-compose env_file
    if os.Getenv("ENVIRONMENT") != "production" {
        _ = godotenv.Load()
    }

    // Load config
    cfg := config.Load()
    logger := config.NewLogger(cfg)
    defer func() { _ = logger.Sync() }()

    // Setup infrastructure
    infra := SetupInfrastructure(cfg, logger)
    defer infra.Close()

    // Setup application
    app := SetupApp(cfg, infra, logger)

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    go func() {
    	logger.Info("Server started", zap.Int("Port", 8080))
        if err := app.Listen(":8080"); err != nil {
        	logger.Fatal("Server failed", zap.Error(err))
        }
    }()

    <-quit
    logger.Info("Received shutdown signal")
    if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
        logger.Error("Server shutdown failed", zap.Error(err))
    } else {
        logger.Info("Server shutdown complete")
    }
}
