package app

import (
	"fmt"
	"log"

	"github.com/fabianoflorentino/mr-robot/adapters/outbound/gateway"
	"github.com/fabianoflorentino/mr-robot/adapters/outbound/persistence/data"
	"github.com/fabianoflorentino/mr-robot/config"
	"github.com/fabianoflorentino/mr-robot/core/services"
	"github.com/fabianoflorentino/mr-robot/database"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
	"gorm.io/gorm"
)

// Container defines the interface for dependency injection container
type Container interface {
	GetDB() *gorm.DB
	GetPaymentService() *services.PaymentService
	GetPaymentQueue() *queue.PaymentQueue
	Shutdown() error
}

// AppContainer implements the Container interface and manages application dependencies
type AppContainer struct {
	config         *config.AppConfig
	dbConnection   database.DatabaseConnection
	db             *gorm.DB
	paymentService *services.PaymentService
	paymentQueue   *queue.PaymentQueue
}

// NewAppContainer creates a new application container with all dependencies initialized
func NewAppContainer() (Container, error) {
	container := &AppContainer{}

	// Step 1: Load configuration
	if err := container.loadConfiguration(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Step 2: Initialize database
	if err := container.initializeDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Step 3: Initialize all services
	if err := container.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Step 4: Run database migrations
	if err := container.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return container, nil
}

// loadConfiguration loads the application configuration
func (c *AppContainer) loadConfiguration() error {
	cfg, err := config.LoadAppConfig()
	if err != nil {
		return err
	}
	c.config = cfg
	return nil
}

// initializeDatabase sets up the database connection
func (c *AppContainer) initializeDatabase() error {
	dbConn, err := database.NewDatabaseConnection(&c.config.Database)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}

	db, err := dbConn.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	c.dbConnection = dbConn
	c.db = db
	return nil
}

// initializeServices sets up all application services
func (c *AppContainer) initializeServices() error {
	// Initialize payment service first (queue depends on it)
	if err := c.initializePaymentService(); err != nil {
		return fmt.Errorf("failed to initialize payment service: %w", err)
	}

	// Initialize payment queue
	if err := c.initializePaymentQueue(); err != nil {
		return fmt.Errorf("failed to initialize payment queue: %w", err)
	}

	return nil
}

// initializePaymentService creates and configures the payment service
func (c *AppContainer) initializePaymentService() error {
	// Create payment repository
	paymentRepo := data.NewDataPaymentRepository(c.db)

	// Create payment processor
	processor := gateway.NewDefaultProcessor(c.config.Payment.DefaultProcessorURL)

	// Create payment service
	c.paymentService = services.NewPaymentService(paymentRepo, &processor)

	return nil
}

// initializePaymentQueue creates and configures the payment queue
func (c *AppContainer) initializePaymentQueue() error {
	c.paymentQueue = queue.NewPaymentQueue(
		c.config.Queue.Workers,
		c.config.Queue.BufferSize,
		c.paymentService,
	)
	return nil
}

// runMigrations executes database migrations
func (c *AppContainer) runMigrations() error {
	if err := c.db.AutoMigrate(&data.Payment{}); err != nil {
		return fmt.Errorf("failed to migrate payment model: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// GetDB returns the database connection
func (c *AppContainer) GetDB() *gorm.DB {
	return c.db
}

// GetPaymentService returns the payment service instance
func (c *AppContainer) GetPaymentService() *services.PaymentService {
	return c.paymentService
}

// GetPaymentQueue returns the payment queue instance
func (c *AppContainer) GetPaymentQueue() *queue.PaymentQueue {
	return c.paymentQueue
}

// Shutdown gracefully shuts down all container components
func (c *AppContainer) Shutdown() error {
	log.Println("Shutting down application container...")

	// Shutdown queue first to stop processing
	if c.paymentQueue != nil {
		log.Println("Shutting down payment queue...")
		c.paymentQueue.Shutdown()
	}

	// Close database connection
	if c.dbConnection != nil {
		log.Println("Closing database connection...")
		if err := c.dbConnection.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	log.Println("Application container shutdown completed successfully")
	return nil
}
