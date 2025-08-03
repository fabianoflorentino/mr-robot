package config

import (
	"fmt"
	"os"
	"strconv"
)

type PaymentConfig struct {
	DefaultProcessorURL  string
	FallbackProcessorURL string
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
	Workers               int
	BufferSize            int
	MaxEnqueueRetries     int
	MaxSimultaneousWrites int
}

func LoadAppConfig() (*AppConfig, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	workers, err := strconv.Atoi(getEnvOrDefault("QUEUE_WORKERS", "50"))
	if err != nil {
		workers = 50
	}

	bufferSize, err := strconv.Atoi(getEnvOrDefault("QUEUE_BUFFER_SIZE", "5000"))
	if err != nil {
		bufferSize = 5000
	}

	maxEnqueueRetries, err := strconv.Atoi(getEnvOrDefault("QUEUE_MAX_ENQUEUE_RETRIES", "3"))
	if err != nil {
		maxEnqueueRetries = 3
	}

	maxSimultaneousWrites, err := strconv.Atoi(getEnvOrDefault("QUEUE_MAX_SIMULTANEOUS_WRITES", "250"))
	if err != nil {
		maxSimultaneousWrites = 250
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
			DefaultProcessorURL:  getEnvOrDefault("DEFAULT_PROCESSOR_URL", ""),
			FallbackProcessorURL: getEnvOrDefault("FALLBACK_PROCESSOR_URL", ""),
		},
		Queue: QueueConfig{
			Workers:               workers,
			BufferSize:            bufferSize,
			MaxEnqueueRetries:     maxEnqueueRetries,
			MaxSimultaneousWrites: maxSimultaneousWrites,
		},
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
