package queue

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds queue-specific configuration
type Config struct {
	Workers               int
	BufferSize            int
	MaxEnqueueRetries     int
	MaxSimultaneousWrites int
}

// ConfigManager manages queue configuration
type ConfigManager struct {
	config *Config
}

// NewConfigManager creates a new queue configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// LoadConfig loads queue configuration from environment variables
func (cm *ConfigManager) LoadConfig() error {
	workers, err := strconv.Atoi(getEnvOrDefault("QUEUE_WORKERS", "10"))
	if err != nil {
		return fmt.Errorf("invalid QUEUE_WORKERS value: %w", err)
	}

	bufferSize, err := strconv.Atoi(getEnvOrDefault("QUEUE_BUFFER_SIZE", "10000"))
	if err != nil {
		return fmt.Errorf("invalid QUEUE_BUFFER_SIZE value: %w", err)
	}

	maxEnqueueRetries, err := strconv.Atoi(getEnvOrDefault("QUEUE_MAX_ENQUEUE_RETRIES", "4"))
	if err != nil {
		return fmt.Errorf("invalid QUEUE_MAX_ENQUEUE_RETRIES value: %w", err)
	}

	maxSimultaneousWrites, err := strconv.Atoi(getEnvOrDefault("QUEUE_MAX_SIMULTANEOUS_WRITES", "50"))
	if err != nil {
		return fmt.Errorf("invalid QUEUE_MAX_SIMULTANEOUS_WRITES value: %w", err)
	}

	cm.config = &Config{
		Workers:               workers,
		BufferSize:            bufferSize,
		MaxEnqueueRetries:     maxEnqueueRetries,
		MaxSimultaneousWrites: maxSimultaneousWrites,
	}

	return nil
}

// GetConfig returns the loaded queue configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetConfig sets the configuration (useful for testing)
func (cm *ConfigManager) SetConfig(config *Config) {
	cm.config = config
}

// Validate validates the queue configuration
func (cm *ConfigManager) Validate() error {
	if cm.config == nil {
		return fmt.Errorf("queue configuration not loaded")
	}

	if cm.config.Workers <= 0 {
		return fmt.Errorf("queue workers must be greater than 0")
	}

	if cm.config.BufferSize <= 0 {
		return fmt.Errorf("queue buffer size must be greater than 0")
	}

	if cm.config.MaxEnqueueRetries < 0 {
		return fmt.Errorf("max enqueue retries cannot be negative")
	}

	if cm.config.MaxSimultaneousWrites <= 0 {
		return fmt.Errorf("max simultaneous writes must be greater than 0")
	}

	return nil
}

// getEnvOrDefault retrieves the value of an environment variable or returns a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
