package container

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/fabianoflorentino/mr-robot/internal/app/config"
	"github.com/fabianoflorentino/mr-robot/internal/app/database"
	"github.com/fabianoflorentino/mr-robot/internal/app/interfaces"
	"github.com/fabianoflorentino/mr-robot/internal/app/migration"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
	appServices "github.com/fabianoflorentino/mr-robot/internal/app/services"
)

// Container defines the interface for dependency injection container
type Container interface {
	GetDB() *sql.DB
	GetPaymentService() interfaces.PaymentServiceInterface
	GetPaymentQueue() *queue.PaymentQueue
	Shutdown() error
}

// AppContainer implements the Container interface and manages application dependencies
type AppContainer struct {
	configManager    *config.Manager
	databaseManager  *database.Manager
	serviceManager   *appServices.Manager
	migrationManager *migration.Manager
}

// NewAppContainer creates a new application container with all dependencies initialized
func NewAppContainer() (Container, error) {
	container := &AppContainer{}

	// Step 1: Initialize configuration manager and load configuration
	container.configManager = config.NewManager()
	if err := container.configManager.LoadConfiguration(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := container.configManager.ValidateConfiguration(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Step 2: Initialize database manager and connect
	container.databaseManager = database.NewManager(container.configManager.GetDatabaseManager())
	if err := container.databaseManager.InitializeDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Step 3: Initialize service manager
	container.serviceManager = appServices.NewManager(
		container.databaseManager.GetDB(),
		container.configManager.GetPaymentConfig(),
		container.configManager.GetQueueConfig(),
		container.configManager.GetCircuitBreakerConfig(),
	)
	if err := container.serviceManager.InitializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Step 4: Initialize migration manager and run migrations
	container.migrationManager = migration.NewManager(container.databaseManager.GetDB())
	if err := container.migrationManager.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return container, nil
}

// GetDB returns the database connection
func (c *AppContainer) GetDB() *sql.DB {
	return c.databaseManager.GetDB()
}

// GetPaymentService returns the payment service instance
func (c *AppContainer) GetPaymentService() interfaces.PaymentServiceInterface {
	return c.serviceManager.GetPaymentService()
}

// GetPaymentQueue returns the payment queue instance
func (c *AppContainer) GetPaymentQueue() *queue.PaymentQueue {
	return c.serviceManager.GetPaymentQueue()
}

// Shutdown gracefully shuts down all container components
func (c *AppContainer) Shutdown() error {
	log.Println("Shutting down application container...")

	// Shutdown services first to stop processing
	if c.serviceManager != nil {
		log.Println("Shutting down services...")
		c.serviceManager.Shutdown()
	}

	// Close database connection
	if c.databaseManager != nil {
		log.Println("Closing database connection...")
		if err := c.databaseManager.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	log.Println("Application container shutdown completed successfully")
	return nil
}
