// cmd/worker/infra.go
package main

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/Ega-telkom/fundivest-backend/internal/config"
	"github.com/Ega-telkom/fundivest-backend/internal/storage"
)

type Infrastructure struct {
	DB          *gorm.DB
	FileStorage storage.FileStorage
	Logger      *zap.Logger
}

func SetupInfrastructure(cfg *config.Config, logger *zap.Logger) *Infrastructure {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("Connected to PostgreSQL")

	// Initialize Storage
	fileStorage, err := storage.NewMinIOStorage(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucket,
		cfg.MinIOUseSSL,
	)
	if err != nil {
		logger.Fatal("Failed to initialize MinIO", zap.Error(err))
	}

	logger.Info("MinIO storage initialized", zap.String("bucket", cfg.MinIOBucket))

	return &Infrastructure{
		DB:          db,
		FileStorage: fileStorage,
	}
}

func (i *Infrastructure) Close() {
	if i.DB != nil {
		sqlDB, _ := i.DB.DB()
		if err := sqlDB.Close(); err != nil {
			i.Logger.Warn("Database close failed", zap.Error(err))
		}
	}
}
