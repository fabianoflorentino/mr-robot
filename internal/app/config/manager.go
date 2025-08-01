package config

import (
	"fmt"

	"github.com/fabianoflorentino/mr-robot/config"
)

// Manager handles configuration loading and management
type Manager struct {
	config *config.AppConfig
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{}
}

// LoadConfiguration loads the application configuration
func (c *Manager) LoadConfiguration() error {
	if cfg, err := config.LoadAppConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	} else {
		c.config = cfg
	}

	return nil
}

// GetConfig returns the loaded configuration
func (c *Manager) GetConfig() *config.AppConfig {
	return c.config
}

// SetConfig sets the configuration (useful for testing)
func (c *Manager) SetConfig(cfg *config.AppConfig) {
	c.config = cfg
}
