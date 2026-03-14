// internal/config/config.go
package config

import (
    "os"
    "time"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Config struct {
    DatabaseURL   string
    ValkeyURL     string
    GotenbergURL  string
    SessionTTL    time.Duration
    
    // MinIO config
    MinIOEndpoint  string
    MinIOAccessKey string
    MinIOSecretKey string
    MinIOBucket    string
    MinIOUseSSL    bool
    
    TemplatePath string
    FrontendURL string
    
    // "development" atau "production"
    Environment string
}

func Load() *Config {
    return &Config{
        DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost/certdb"),
        ValkeyURL:    getEnv("VALKEY_URL", "localhost:6379"),
        GotenbergURL: getEnv("GOTENBERG_URL", "http://localhost:3000"),
        SessionTTL:   24 * time.Hour,
        
        MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
        MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
        MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
        MinIOBucket:    getEnv("MINIO_BUCKET", "certificates"),
        MinIOUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
        
        TemplatePath: getEnv("TEMPLATE_PATH", "./templates/certificate.html"),
        FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
        
        Environment: getEnv("ENVIRONMENT", "development"),
    }
}

func NewLogger(env string) *zap.Logger {
    var logger *zap.Logger
    var err error
    
    if env == "production" {
        // Production: JSON output
        config := zap.NewProductionConfig()
        config.EncoderConfig.TimeKey = "timestamp"
        config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
        logger, err = config.Build()
    } else {
        // Development: Console output with colors
        config := zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
        logger, err = config.Build()
    }
    
    if err != nil {
        panic(err)
    }
    
    return logger
}

func getEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}