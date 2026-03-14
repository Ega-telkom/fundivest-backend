// cmd/api/main.go
package main

import (
	"log"
	"os"
	"time"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"fundivest/internal/config"
)

// @title           fundivest REST API
// @version         1.0
// @description     Manajemen Sesi & Sertifikat untuk fundivest
//
// @host      localhost:8080
// @BasePath  /api/v1
func main() {
    // Load environment
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // Load config
    cfg := config.Load()
    logger := config.NewLogger(cfg.Environment)
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
