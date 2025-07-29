package app

import (
	"fmt"

	"github.com/fabianoflorentino/mr-robot/config"
	"github.com/fabianoflorentino/mr-robot/database"
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

// WithDefaultConfig sets the default application configuration if none is provided
func (b *ContainerBuilder) Build() (Container, error) {
	if b.config == nil {
		cfg, err := config.LoadAppConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load default configuration: %w", err)
		}
		b.config = cfg
	}

	if b.dbConnection == nil {
		conn, err := database.NewDatabaseConnection(&b.config.Database)
		if err != nil {
			return nil, fmt.Errorf("failed to create default database connection: %w", err)
		}
		b.dbConnection = conn
	}

	db, err := b.dbConnection.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	container := &AppContainer{
		config:       b.config,
		dbConnection: b.dbConnection,
		db:           db,
	}

	if err := container.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	return container, nil
}
