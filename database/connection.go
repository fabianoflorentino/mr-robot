package database

import (
	"database/sql"
	"fmt"

	"github.com/fabianoflorentino/mr-robot/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DatabaseConnection interface {
	Connect() (*sql.DB, error)
	Close() error
	GetDB() *sql.DB
}

type PostgresConnection struct {
	config *config.DatabaseConfig
	db     *sql.DB
}

func NewPostgresConnection(cfg *config.DatabaseConfig) *PostgresConnection {
	return &PostgresConnection{
		config: cfg,
	}
}

func (p *PostgresConnection) Connect() (*sql.DB, error) {
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

func (p *PostgresConnection) Close() error {
	if p.db == nil {
		return nil
	}

	return p.db.Close()
}

func (p *PostgresConnection) GetDB() *sql.DB {
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
