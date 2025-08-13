package database

import (
	"database/sql"
	"fmt"

	"github.com/fabianoflorentino/mr-robot/database"
)

// Manager handles database-related operations
type Manager struct {
	configManager *ConfigManager
	dbConnection  database.DatabaseConnection
	db            *sql.DB
}

// NewManager creates a new database manager
func NewManager(configManager *ConfigManager) *Manager {
	return &Manager{
		configManager: configManager,
	}
}

// InitializeDatabase sets up the database connection
func (d *Manager) InitializeDatabase() error {
	if d.configManager == nil {
		return fmt.Errorf("database config manager not set")
	}

	config := d.configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("database configuration not loaded")
	}

	// Convert to database package format
	dbConfig := &database.DatabaseConfig{
		Host:     config.Host,
		Port:     config.Port,
		User:     config.User,
		Password: config.Password,
		Database: config.Database,
		SSLMode:  config.SSLMode,
		Timezone: config.Timezone,
	}

	if dbConn, err := database.NewDatabaseConnection(dbConfig); err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	} else {
		db, err := dbConn.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		d.dbConnection = dbConn
		d.db = db
	}

	return nil
}

// GetDB returns the database connection
func (d *Manager) GetDB() *sql.DB {
	return d.db
}

// GetConnection returns the database connection manager
func (d *Manager) GetConnection() database.DatabaseConnection {
	return d.dbConnection
}

// GetConfigManager returns the database configuration manager
func (d *Manager) GetConfigManager() *ConfigManager {
	return d.configManager
}

// Close closes the database connection
func (d *Manager) Close() error {
	if d.dbConnection != nil {
		return d.dbConnection.Close()
	}

	return nil
}
