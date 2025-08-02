package config

import (
	"fmt"
	"os"
	"strconv"
)

type PaymentConfig struct {
	DefaultProcessorURL string
}

type AppConfig struct {
	Database DatabaseConfig
	Payment  PaymentConfig
	Queue    QueueConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	Timezone string
}

type QueueConfig struct {
	Workers    int
	BufferSize int
}

func LoadAppConfig() (*AppConfig, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	workers, err := strconv.Atoi(getEnvOrDefault("QUEUE_WORKERS", "4"))
	if err != nil {
		workers = 4
	}

	bufferSize, err := strconv.Atoi(getEnvOrDefault("QUEUE_BUFFER_SIZE", "100"))
	if err != nil {
		bufferSize = 100
	}

	return &AppConfig{
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
			Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
			User:     getEnvOrDefault("POSTGRES_USER", "postgres"),
			Password: getEnvOrDefault("POSTGRES_PASSWORD", ""),
			Database: getEnvOrDefault("POSTGRES_DB", "mr_robot"),
			SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "disable"),
			Timezone: getEnvOrDefault("POSTGRES_TIMEZONE", "UTC"),
		},
		Payment: PaymentConfig{
			DefaultProcessorURL: getEnvOrDefault("DEFAULT_PROCESSOR_URL", ""),
		},
		Queue: QueueConfig{
			Workers:    workers,
			BufferSize: bufferSize,
		},
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
