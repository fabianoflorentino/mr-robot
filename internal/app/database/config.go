package database

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds database-specific configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	Timezone string
}

// ConfigManager manages database configuration
type ConfigManager struct {
	config *Config
}

// NewConfigManager creates a new database configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// LoadConfig loads database configuration from environment variables
func (cm *ConfigManager) LoadConfig() error {
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")
	user := getEnvOrDefault("POSTGRES_USER", "postgres")
	password := os.Getenv("POSTGRES_PASSWORD")
	database := getEnvOrDefault("POSTGRES_DB", "mr_robot")
	sslMode := getEnvOrDefault("POSTGRES_SSLMODE", "disable")
	timezone := getEnvOrDefault("POSTGRES_TIMEZONE", "UTC")

	cm.config = &Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		SSLMode:  sslMode,
		Timezone: timezone,
	}

	return nil
}

// GetConfig returns the loaded database configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetConfig sets the configuration (useful for testing)
func (cm *ConfigManager) SetConfig(config *Config) {
	cm.config = config
}

// Validate validates the database configuration
func (cm *ConfigManager) Validate() error {
	if cm.config == nil {
		return fmt.Errorf("database configuration not loaded")
	}

	if cm.config.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}

	if cm.config.Port == "" {
		return fmt.Errorf("database port cannot be empty")
	}

	// Validate port is a number
	if _, err := strconv.Atoi(cm.config.Port); err != nil {
		return fmt.Errorf("invalid database port: %w", err)
	}

	if cm.config.User == "" {
		return fmt.Errorf("database user cannot be empty")
	}

	if cm.config.Database == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	// Validate SSL mode
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
	isValidSSL := false
	for _, mode := range validSSLModes {
		if cm.config.SSLMode == mode {
			isValidSSL = true
			break
		}
	}
	if !isValidSSL {
		return fmt.Errorf("invalid SSL mode: %s. Valid modes are: %v", cm.config.SSLMode, validSSLModes)
	}

	return nil
}

// GetConnectionString returns the database connection string
func (cm *ConfigManager) GetConnectionString() string {
	if cm.config == nil {
		return ""
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s TimeZone=%s",
		cm.config.Host,
		cm.config.Port,
		cm.config.User,
		cm.config.Database,
		cm.config.SSLMode,
		cm.config.Timezone,
	)

	if cm.config.Password != "" {
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			cm.config.Host,
			cm.config.Port,
			cm.config.User,
			cm.config.Password,
			cm.config.Database,
			cm.config.SSLMode,
			cm.config.Timezone,
		)
	}

	return connStr
}

// getEnvOrDefault retrieves the value of an environment variable or returns a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
