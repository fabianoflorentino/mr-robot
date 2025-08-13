package config

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ControllerConfig struct {
	ContentType, ApplicationJSON, HostName, TimeInfo string
	StatusOK                                         int
	TimeAfter                                        time.Duration
}

type PaymentConfig struct {
	DefaultProcessorURL, FallbackProcessorURL string
}

type DatabaseConfig struct {
	Host, Port, User, Password, Database, SSLMode, Timezone string
}

type QueueConfig struct {
	Workers, BufferSize, MaxEnqueueRetries, MaxSimultaneousWrites int
}

type CircuitBreakerConfig struct {
	Timeout, ResetTimeout  time.Duration
	MaxFailures, RateLimit int
}

type AppConfig struct {
	Database         DatabaseConfig
	Payment          PaymentConfig
	Queue            QueueConfig
	CircuitBreaker   CircuitBreakerConfig
	ControllerConfig ControllerConfig
}

// LoadEnv loads environment variables from a .env file if it exists.
func LoadAppConfig() (*AppConfig, error) {
	if err := LoadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	// Queue configuration
	workers, _ := strconv.Atoi(getEnvOrDefault("QUEUE_WORKERS", "10"))
	bufferSize, _ := strconv.Atoi(getEnvOrDefault("QUEUE_BUFFER_SIZE", "10000"))
	maxEnqueueRetries, _ := strconv.Atoi(getEnvOrDefault("QUEUE_MAX_ENQUEUE_RETRIES", "4"))
	maxSimultaneousWrites, _ := strconv.Atoi(getEnvOrDefault("QUEUE_MAX_SIMULTANEOUS_WRITES", "50"))

	// Circuit Breaker configuration
	circuitBreakerTimeout, _ := time.ParseDuration(getEnvOrDefault("CIRCUIT_BREAKER_TIMEOUT", "1s"))
	circuitBreakerMaxFailures, _ := strconv.Atoi(getEnvOrDefault("CIRCUIT_BREAKER_MAX_FAILURES", "5"))
	circuitBreakerResetTimeout, _ := time.ParseDuration(getEnvOrDefault("CIRCUIT_BREAKER_RESET_TIMEOUT", "10s"))
	circuitBreakerRateLimit, _ := strconv.Atoi(getEnvOrDefault("CIRCUIT_BREAKER_RATE_LIMIT", "5"))

	// Database configuration
	databaseHost := getEnvOrDefault("POSTGRES_HOST", "localhost")
	databasePort := getEnvOrDefault("POSTGRES_PORT", "5432")
	databaseUser := getEnvOrDefault("POSTGRES_USER", "postgres")
	databasePassword := getEnvOrDefault("POSTGRES_PASSWORD", "")
	databaseName := getEnvOrDefault("POSTGRES_DB", "mr_robot")
	databaseSSLMode := getEnvOrDefault("POSTGRES_SSLMODE", "disable")
	databaseTimezone := getEnvOrDefault("POSTGRES_TIMEZONE", "UTC")

	// Payment configuration
	defaultProcessorURL := getEnvOrDefault("DEFAULT_PROCESSOR_URL", "")
	fallbackProcessorURL := getEnvOrDefault("FALLBACK_PROCESSOR_URL", "")

	// Controller configuration
	controllerContentType := "Content-Type"
	controllerApplicationJSON := "application/json"
	controllerHostName := getEnvOrDefault("HOSTNAME", "localhost")
	controllerStatusOK := http.StatusOK
	controllerTimeInfo := time.Now().Format(time.RFC3339)
	controllerTimeAfter := 250 * time.Millisecond

	return &AppConfig{
		Database: DatabaseConfig{
			Host:     databaseHost,
			Port:     databasePort,
			User:     databaseUser,
			Password: databasePassword,
			Database: databaseName,
			SSLMode:  databaseSSLMode,
			Timezone: databaseTimezone,
		},
		Payment: PaymentConfig{
			DefaultProcessorURL:  defaultProcessorURL,
			FallbackProcessorURL: fallbackProcessorURL,
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
		ControllerConfig: ControllerConfig{
			ContentType:     controllerContentType,
			ApplicationJSON: controllerApplicationJSON,
			HostName:        controllerHostName,
			StatusOK:        controllerStatusOK,
			TimeInfo:        controllerTimeInfo,
			TimeAfter:       controllerTimeAfter,
		},
	}, nil
}

// getEnvOrDefault retrieves the value of an environment variable or returns a default value if not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
