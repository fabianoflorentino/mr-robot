package circuitbreaker

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds circuit breaker configuration
type Config struct {
	Timeout      time.Duration
	ResetTimeout time.Duration
	MaxFailures  int
	RateLimit    int
}

// ConfigManager manages circuit breaker configuration
type ConfigManager struct {
	config *Config
}

// NewConfigManager creates a new circuit breaker configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// LoadConfig loads circuit breaker configuration from environment variables
func (cm *ConfigManager) LoadConfig() error {
	timeout, err := time.ParseDuration(getEnvOrDefault("CIRCUIT_BREAKER_TIMEOUT", "1s"))
	if err != nil {
		return fmt.Errorf("invalid CIRCUIT_BREAKER_TIMEOUT value: %w", err)
	}

	resetTimeout, err := time.ParseDuration(getEnvOrDefault("CIRCUIT_BREAKER_RESET_TIMEOUT", "10s"))
	if err != nil {
		return fmt.Errorf("invalid CIRCUIT_BREAKER_RESET_TIMEOUT value: %w", err)
	}

	maxFailures, err := strconv.Atoi(getEnvOrDefault("CIRCUIT_BREAKER_MAX_FAILURES", "5"))
	if err != nil {
		return fmt.Errorf("invalid CIRCUIT_BREAKER_MAX_FAILURES value: %w", err)
	}

	rateLimit, err := strconv.Atoi(getEnvOrDefault("CIRCUIT_BREAKER_RATE_LIMIT", "5"))
	if err != nil {
		return fmt.Errorf("invalid CIRCUIT_BREAKER_RATE_LIMIT value: %w", err)
	}

	cm.config = &Config{
		Timeout:      timeout,
		ResetTimeout: resetTimeout,
		MaxFailures:  maxFailures,
		RateLimit:    rateLimit,
	}

	return nil
}

// GetConfig returns the loaded circuit breaker configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetConfig sets the configuration (useful for testing)
func (cm *ConfigManager) SetConfig(config *Config) {
	cm.config = config
}

// Validate validates the circuit breaker configuration
func (cm *ConfigManager) Validate() error {
	if cm.config == nil {
		return fmt.Errorf("circuit breaker configuration not loaded")
	}

	if cm.config.Timeout <= 0 {
		return fmt.Errorf("circuit breaker timeout must be greater than 0")
	}

	if cm.config.ResetTimeout <= 0 {
		return fmt.Errorf("circuit breaker reset timeout must be greater than 0")
	}

	if cm.config.MaxFailures <= 0 {
		return fmt.Errorf("circuit breaker max failures must be greater than 0")
	}

	if cm.config.RateLimit <= 0 {
		return fmt.Errorf("circuit breaker rate limit must be greater than 0")
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
