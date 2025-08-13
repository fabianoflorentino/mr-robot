package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	Timezone string
}

type DatabaseConnection interface {
	Connect() (*sql.DB, error)
	Close() error
	GetDB() *sql.DB
}

type DatabaseConfiguration struct {
	config *DatabaseConfig
	db     *sql.DB
}

// NewDatabaseConnection creates a new database connection with the provided config
func NewDatabaseConnection(cfg *DatabaseConfig) (DatabaseConnection, error) {
	return &DatabaseConfiguration{config: cfg}, nil
}

func (p *DatabaseConfiguration) Connect() (*sql.DB, error) {
	dsn := p.buildConnectionString()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	p.db = db
	return db, nil
}

func (p *DatabaseConfiguration) Close() error {
	if p.db == nil {
		return nil
	}

	return p.db.Close()
}

func (p *DatabaseConfiguration) GetDB() *sql.DB {
	return p.db
}

func (p *DatabaseConfiguration) buildConnectionString() string {
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
