package database

import (
	"fmt"

	"github.com/fabianoflorentino/mr-robot/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConnection interface {
	Connect() (*gorm.DB, error)
	Close() error
	GetDB() *gorm.DB
}

type PostgresConnection struct {
	config *config.DatabaseConfig
	db     *gorm.DB
}

func NewPostgresConnection(cfg *config.DatabaseConfig) *PostgresConnection {
	return &PostgresConnection{
		config: cfg,
	}
}

func (p *PostgresConnection) Connect() (*gorm.DB, error) {
	dsn := p.buildConnectionString()

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), gormConfig)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	p.db = db
	return db, nil
}

func (p *PostgresConnection) Close() error {
	if p.db == nil {
		return nil
	}

	sqlDB, err := p.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	return sqlDB.Close()
}

func (p *PostgresConnection) GetDB() *gorm.DB {
	return p.db
}

func (p *PostgresConnection) buildConnectionString() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		p.config.Host,
		p.config.User,
		p.config.Password,
		p.config.Database,
		p.config.Port,
		p.config.SSLMode,
		p.config.Timezone,
	)
}

// NewDatabaseConnection creates a new database connection with the provided config
func NewDatabaseConnection(cfg *config.DatabaseConfig) (DatabaseConnection, error) {
	return NewPostgresConnection(cfg), nil
}
