package migration

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	databaseName = os.Getenv("POSTGRES_DB")
)

// Manager handles database migrations
type Manager struct {
	db    *sql.DB
	mutex sync.Mutex
}

// NewManager creates a new migration manager
func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db:    db,
		mutex: sync.Mutex{},
	}
}

// RunMigrations executes database migrations using native SQL
func (m *Manager) RunMigrations() error {
	// Use mutex to prevent concurrent migration execution
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if database exists
	if !m.isDatabaseExists(databaseName) {
		return fmt.Errorf("database %q does not exist", databaseName)
	}

	// Check if payments table already exists
	if !m.isTableExists("payments") {
		if err := m.createPaymentsTable(); err != nil {
			return fmt.Errorf("failed to create payments table: %w", err)
		}

		log.Println("Payments table created successfully")
	} else {
		log.Println("Payments table already exists, skipping migration")
	}

	log.Println("Database migrations completed successfully")

	return nil
}

func (m *Manager) isDatabaseExists(database string) bool {
	var exists bool

	err := m.db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", database).Scan(&exists)

	if err != nil {
		return false
	}

	return exists
}

func (m *Manager) isTableExists(tableName string) bool {
	var exists bool

	query := `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = $1
	)`

	err := m.db.QueryRow(query, tableName).Scan(&exists)

	if err != nil {
		return false
	}

	return exists
}

func (m *Manager) createPaymentsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS payments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		correlation_id UUID NOT NULL,
		amount DECIMAL(15,2) NOT NULL,
		processor VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_payments_correlation_id ON payments(correlation_id);
	CREATE INDEX IF NOT EXISTS idx_payments_processor ON payments(processor);
	CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);
	`

	_, err := m.db.Exec(query)
	return err
}
