// cmd/api/infra.go
package main

import (
	"time"

	"github.com/valkey-io/valkey-go"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"fundivest/internal/config"
	"fundivest/internal/queue"
	repoPostgres "fundivest/internal/repository/postgres"
	"fundivest/internal/storage"
)

type Infrastructure struct {
	DB             *gorm.DB
	ValkeyClient   valkey.Client
	QueuePublisher *queue.AsynqPublisher
	FileStorage    storage.FileStorage
	Logger         *zap.Logger
}

func SetupInfrastructure(cfg *config.Config, logger *zap.Logger) *Infrastructure {
	// Init Postgres
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Auto migrate
	if err := db.AutoMigrate(&repoPostgres.Certificate{}); err != nil {
		logger.Fatal("Failed to migrate", zap.Error(err))
	}
	logger.Info("Connected to PostgreSQL")

	// Init Valkey
	valkeyClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.ValkeyURL},
	})
	if err != nil {
		logger.Fatal("Failed to connect to Valkey", zap.Error(err))
	}
	logger.Info("Connected to Valkey")

	// Init Queue
	queuePublisher := queue.NewAsynqPublisher(cfg.ValkeyURL)
	logger.Info("Queue publisher initialized")

	// Init Storage
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
		DB:             db,
		ValkeyClient:   valkeyClient,
		QueuePublisher: queuePublisher,
		FileStorage:    fileStorage,
	}
}

func (i *Infrastructure) Close() {
	i.Logger.Info("Shutting down infrastructure...")

	if i.ValkeyClient != nil {
		i.ValkeyClient.Close()
		i.Logger.Info("Valkey closed")
	}

	if i.QueuePublisher != nil {
		if err := i.QueuePublisher.Close(); err != nil {
			i.Logger.Warn("Queue close failed", zap.Error(err))
		} else {
			i.Logger.Info("Queue closed")
		}
	}

	if i.DB != nil {
		sqlDB, _ := i.DB.DB()
		if err := sqlDB.Close(); err != nil {
			i.Logger.Warn("Database close failed", zap.Error(err))
		} else {
			i.Logger.Info("Database closed")
		}
	}

	i.Logger.Info("Shutdown complete")
	_ = i.Logger.Sync()
}
