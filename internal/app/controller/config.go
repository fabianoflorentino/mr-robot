package controller

import (
	"fmt"
	"os"
	"time"
)

// Config holds controller-specific configuration
type Config struct {
	ContentType     string
	ApplicationJSON string
	HostName        string
	TimeInfo        string
	StatusOK        int
	TimeAfter       time.Duration
}

// ConfigManager manages controller configuration
type ConfigManager struct {
	config *Config
}

// NewConfigManager creates a new controller configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// LoadConfig loads controller configuration from environment variables
func (cm *ConfigManager) LoadConfig() error {
	hostName := getEnvOrDefault("HOSTNAME", "localhost")

	cm.config = &Config{
		ContentType:     "Content-Type",
		ApplicationJSON: "application/json",
		HostName:        hostName,
		TimeAfter:       250 * time.Millisecond,
	}

	return nil
}

// GetConfig returns the loaded controller configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetConfig sets the configuration (useful for testing)
func (cm *ConfigManager) SetConfig(config *Config) {
	cm.config = config
}

// Validate validates the controller configuration
func (cm *ConfigManager) Validate() error {
	if cm.config == nil {
		return fmt.Errorf("controller configuration not loaded")
	}

	if cm.config.HostName == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	if cm.config.TimeAfter <= 0 {
		return fmt.Errorf("time after must be greater than 0")
	}

	return nil
}

// RefreshTimeInfo updates the time info to current time
func (cm *ConfigManager) RefreshTimeInfo() {
	if cm.config != nil {
		cm.config.TimeInfo = time.Now().Format(time.RFC3339)
	}
}

// getEnvOrDefault retrieves the value of an environment variable or returns a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
