package container

import (
	"fmt"

	"github.com/fabianoflorentino/mr-robot/database"
	appConfig "github.com/fabianoflorentino/mr-robot/internal/app/config"
	appDB "github.com/fabianoflorentino/mr-robot/internal/app/database"
	"github.com/fabianoflorentino/mr-robot/internal/app/migration"
	appServices "github.com/fabianoflorentino/mr-robot/internal/app/services"
)

type ContainerBuilder struct {
	configManager *appConfig.Manager
	dbConnection  database.DatabaseConnection
}

// NewContainerBuilder creates a new instance of ContainerBuilder
func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{}
}

// WithConfigManager sets the configuration manager for the container
func (b *ContainerBuilder) WithConfigManager(cm *appConfig.Manager) *ContainerBuilder {
	b.configManager = cm
	return b
}

// WithDatabaseConnection sets the database connection for the container
func (b *ContainerBuilder) WithDatabaseConnection(conn database.DatabaseConnection) *ContainerBuilder {
	b.dbConnection = conn
	return b
}

// Build creates the container with all managers properly initialized
func (b *ContainerBuilder) Build() (Container, error) {
	// Use provided config manager or create default
	var configManager *appConfig.Manager
	if b.configManager != nil {
		configManager = b.configManager
	} else {
		configManager = appConfig.NewManager()

		// Load configuration
		if err := configManager.LoadConfiguration(); err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}

		// Validate configuration
		if err := configManager.ValidateConfiguration(); err != nil {
			return nil, fmt.Errorf("invalid configuration: %w", err)
		}
	}

	// Create database manager
	databaseManager := appDB.NewManager(configManager.GetDatabaseManager())

	// Use provided database connection or create default
	if b.dbConnection != nil {
		if err := databaseManager.InitializeDatabase(); err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	// Create service manager
	serviceManager := appServices.NewManager(
		databaseManager.GetDB(),
		configManager.GetPaymentConfig(),
		configManager.GetQueueConfig(),
		configManager.GetCircuitBreakerConfig(),
	)

	if err := serviceManager.InitializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

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
