package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type PaymentConfig struct {
	DefaultProcessorURL  string
	FallbackProcessorURL string
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

type CircuitBreakerConfig struct {
	Timeout      time.Duration
	MaxFailures  int
	ResetTimeout time.Duration
	RateLimit    int
}

type AppConfig struct {
	Database       DatabaseConfig
	Payment        PaymentConfig
	Queue          QueueConfig
	CircuitBreaker CircuitBreakerConfig
}

// LoadEnv loads environment variables from a .env file if it exists.
func LoadAppConfig() (*AppConfig, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	workers, err := strconv.Atoi(getEnvOrDefault("QUEUE_WORKERS", "10"))
	if err != nil {
		workers = 10
	}

	bufferSize, err := strconv.Atoi(getEnvOrDefault("QUEUE_BUFFER_SIZE", "10000"))
	if err != nil {
		bufferSize = 10000
	}

	maxEnqueueRetries, err := strconv.Atoi(getEnvOrDefault("QUEUE_MAX_ENQUEUE_RETRIES", "4"))
	if err != nil {
		maxEnqueueRetries = 4
	}

	maxSimultaneousWrites, err := strconv.Atoi(getEnvOrDefault("QUEUE_MAX_SIMULTANEOUS_WRITES", "50"))
	if err != nil {
		maxSimultaneousWrites = 50
	}

	circuitBreakerTimeout, err := time.ParseDuration(getEnvOrDefault("CIRCUIT_BREAKER_TIMEOUT", "1s"))
	if err != nil {
		circuitBreakerTimeout = 1 * time.Second
	}

	circuitBreakerMaxFailures, err := strconv.Atoi(getEnvOrDefault("CIRCUIT_BREAKER_MAX_FAILURES", "5"))
	if err != nil {
		circuitBreakerMaxFailures = 5
	}

	circuitBreakerResetTimeout, err := time.ParseDuration(getEnvOrDefault("CIRCUIT_BREAKER_RESET_TIMEOUT", "10s"))
	if err != nil {
		circuitBreakerResetTimeout = 10 * time.Second
	}

	circuitBreakerRateLimit, err := strconv.Atoi(getEnvOrDefault("CIRCUIT_BREAKER_RATE_LIMIT", "5"))
	if err != nil {
		circuitBreakerRateLimit = 5
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
		CircuitBreaker: CircuitBreakerConfig{
			Timeout:      circuitBreakerTimeout,
			MaxFailures:  circuitBreakerMaxFailures,
			ResetTimeout: circuitBreakerResetTimeout,
			RateLimit:    circuitBreakerRateLimit,
		},
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
