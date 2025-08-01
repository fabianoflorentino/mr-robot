package app

import (
	"fmt"

	"github.com/fabianoflorentino/mr-robot/config"
	"github.com/fabianoflorentino/mr-robot/database"
	appConfig "github.com/fabianoflorentino/mr-robot/internal/app/config"
	appDB "github.com/fabianoflorentino/mr-robot/internal/app/database"
	"github.com/fabianoflorentino/mr-robot/internal/app/migration"
	appServices "github.com/fabianoflorentino/mr-robot/internal/app/services"
)

type ContainerBuilder struct {
	config       *config.AppConfig
	dbConnection database.DatabaseConnection
}

// NewContainerBuilder creates a new instance of ContainerBuilder
func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{}
}

// WithConfig sets the application configuration for the container
func (b *ContainerBuilder) WithConfig(cfg *config.AppConfig) *ContainerBuilder {
	b.config = cfg
	return b
}

// WithDatabaseConnection sets the database connection for the container
func (b *ContainerBuilder) WithDatabaseConnection(conn database.DatabaseConnection) *ContainerBuilder {
	b.dbConnection = conn
	return b
}

// Build creates the container with all managers properly initialized
func (b *ContainerBuilder) Build() (Container, error) {
	// Create configuration manager
	configManager := appConfig.NewManager()

	// Use provided config or load default
	if b.config != nil {
		configManager.SetConfig(b.config)
	} else {
		if err := configManager.LoadConfiguration(); err != nil {
			return nil, fmt.Errorf("failed to load default configuration: %w", err)
		}
	}

	// Create database manager
	databaseManager := appDB.NewManager(configManager.GetConfig())

	// Use provided database connection or create default
	if b.dbConnection != nil {
		// Custom implementation for setting existing connection
		// This would require extending the database manager
		if err := databaseManager.InitializeDatabase(); err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
	} else {
		if err := databaseManager.InitializeDatabase(); err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	// Create service manager
	serviceManager := appServices.NewManager(
		configManager.GetConfig(),
		databaseManager.GetDB(),
	)
	if err := serviceManager.InitializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Create migration manager and run migrations
	migrationManager := migration.NewManager(databaseManager.GetDB())
	if err := migrationManager.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create container with all managers
	container := &AppContainer{
		configManager:    configManager,
		databaseManager:  databaseManager,
		serviceManager:   serviceManager,
		migrationManager: migrationManager,
	}

	return container, nil
}
