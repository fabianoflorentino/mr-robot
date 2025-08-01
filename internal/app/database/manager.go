package database

import (
	"fmt"

	"github.com/fabianoflorentino/mr-robot/config"
	"github.com/fabianoflorentino/mr-robot/database"
	"gorm.io/gorm"
)

// Manager handles database-related operations
type Manager struct {
	config       *config.AppConfig
	dbConnection database.DatabaseConnection
	db           *gorm.DB
}

// NewManager creates a new database manager
func NewManager(cfg *config.AppConfig) *Manager {
	return &Manager{
		config: cfg,
	}
}

// InitializeDatabase sets up the database connection
func (d *Manager) InitializeDatabase() error {
	if dbConn, err := database.NewDatabaseConnection(&d.config.Database); err != nil {
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
func (d *Manager) GetDB() *gorm.DB {
	return d.db
}

// GetConnection returns the database connection manager
func (d *Manager) GetConnection() database.DatabaseConnection {
	return d.dbConnection
}

// Close closes the database connection
func (d *Manager) Close() error {
	if d.dbConnection != nil {
		return d.dbConnection.Close()
	}

	return nil
}
