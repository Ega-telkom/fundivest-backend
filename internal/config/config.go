// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Environment string

	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// Valkey
	ValkeyHost     string
	ValkeyPort     string
	ValkeyPassword string

	// MinIO config
	MinIOHost         string
	MinIOPort         string
	MinIORootUser     string
	MinIORootPassword string
	MinIOBucket       string
	MinIOUseSSL       bool

	// App
	TemplatePath   string
	AllowedOrigins string
	GotenbergURL   string
	SessionTTL     time.Duration
}

func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "certdb"),
		DBUser:     getEnv("DB_USER", "certuser"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Valkey
		ValkeyHost:     getEnv("VALKEY_HOST", "localhost"),
		ValkeyPort:     getEnv("VALKEY_PORT", "6379"),
		ValkeyPassword: getEnv("VALKEY_PASSWORD", ""),

		// MinIO
		MinIOHost:         getEnv("MINIO_HOST", "localhost"),
		MinIOPort:         getEnv("MINIO_PORT", "9000"),
		MinIORootUser:     getEnv("MINIO_ROOT_USER", "minioadmin"),
		MinIORootPassword: getEnv("MINIO_ROOT_PASSWORD", "minioadmin123"),
		MinIOBucket:       getEnv("MINIO_BUCKET", "certificates"),
		MinIOUseSSL:       getEnv("MINIO_USE_SSL", "false") == "true",

		// App
		TemplatePath:   getEnv("TEMPLATE_PATH", "./templates/certificate.html"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		GotenbergURL:   getEnv("GOTENBERG_URL", "http://localhost:3000"),
		SessionTTL:     24 * time.Hour,
	}
}

func NewLogger(cfg *Config) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.IsProduction() {
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

// PostgreSQL address
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

// Valkey address
func (c *Config) ValkeyAddr() string {
	return fmt.Sprintf("%s:%s", c.ValkeyHost, c.ValkeyPort)
}

// MinIO address
func (c *Config) MinioAddr() string {
	return fmt.Sprintf("%s:%s", c.MinIOHost, c.MinIOPort)
}

// Environments
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
