package app

import (
	"database/sql"

	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
)

// ContainerInterface defines the interface for dependency injection container
type ContainerInterface interface {
	GetDB() *sql.DB
	GetPaymentQueue() *queue.PaymentQueue
	Shutdown() error
}

// Initializer defines the interface for component initialization
type Initializer interface {
	Initialize() error
}

// ConfigurationLoader defines the interface for loading configuration
type ConfigurationLoader interface {
	LoadConfiguration() error
}

// DatabaseInitializer defines the interface for database initialization
type DatabaseInitializer interface {
	InitializeDatabase() error
}

// ServiceInitializer defines the interface for service initialization
type ServiceInitializer interface {
	InitializeServices() error
}

// MigrationRunner defines the interface for running database migrations
type MigrationRunner interface {
	RunMigrations() error
}
