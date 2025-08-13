package payment

import (
	"fmt"
	"net/url"
	"os"
)

// Config holds payment processor configuration
type Config struct {
	DefaultProcessorURL  string
	FallbackProcessorURL string
}

// ConfigManager manages payment configuration
type ConfigManager struct {
	config *Config
}

// NewConfigManager creates a new payment configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// LoadConfig loads payment configuration from environment variables
func (cm *ConfigManager) LoadConfig() error {
	defaultProcessorURL := os.Getenv("DEFAULT_PROCESSOR_URL")
	if defaultProcessorURL == "" {
		return fmt.Errorf("DEFAULT_PROCESSOR_URL environment variable is required")
	}

	fallbackProcessorURL := os.Getenv("FALLBACK_PROCESSOR_URL")
	if fallbackProcessorURL == "" {
		return fmt.Errorf("FALLBACK_PROCESSOR_URL environment variable is required")
	}

	cm.config = &Config{
		DefaultProcessorURL:  defaultProcessorURL,
		FallbackProcessorURL: fallbackProcessorURL,
	}

	return nil
}

// GetConfig returns the loaded payment configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetConfig sets the configuration (useful for testing)
func (cm *ConfigManager) SetConfig(config *Config) {
	cm.config = config
}

// Validate validates the payment configuration
func (cm *ConfigManager) Validate() error {
	if cm.config == nil {
		return fmt.Errorf("payment configuration not loaded")
	}

	if cm.config.DefaultProcessorURL == "" {
		return fmt.Errorf("default processor URL cannot be empty")
	}

	if cm.config.FallbackProcessorURL == "" {
		return fmt.Errorf("fallback processor URL cannot be empty")
	}

	// Validate URL format
	if _, err := url.Parse(cm.config.DefaultProcessorURL); err != nil {
		return fmt.Errorf("invalid default processor URL: %w", err)
	}

	if _, err := url.Parse(cm.config.FallbackProcessorURL); err != nil {
		return fmt.Errorf("invalid fallback processor URL: %w", err)
	}

	return nil
}
