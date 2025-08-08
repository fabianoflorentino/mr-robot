package migration

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/fabianoflorentino/mr-robot/adapters/outbound/persistence/data"
	"gorm.io/gorm"
)

var (
	databaseName = os.Getenv("POSTGRES_DB")
)

// Manager handles database migrations
type Manager struct {
	db    *gorm.DB
	mutex sync.Mutex
}

// NewManager creates a new migration manager
func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:    db,
		mutex: sync.Mutex{},
	}
}

// RunMigrations executes database migrations using GORM's built-in capabilities
func (m *Manager) RunMigrations() error {
	// Use mutex to prevent concurrent migration execution
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if database exists
	if !m.isDatabaseExists(databaseName) {
		return fmt.Errorf("database %q does not exist", databaseName)
	}

	// Check if payments table already exists
	if !m.db.Migrator().HasTable(&data.Payment{}) {
		if err := m.db.AutoMigrate(&data.Payment{}); err != nil {
			return fmt.Errorf("failed to migrate payment model: %w", err)
		}

		log.Println("Payments table created successfully")
	} else {
		log.Println("Payments table already exists, skipping migration")
		return nil
	}

	log.Println("Database migrations completed successfully")

	return nil
}

func (m *Manager) isDatabaseExists(database string) bool {
	var exists bool

	err := m.db.Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", database).Scan(&exists).Error

	if err != nil {
		return false
	}

	return exists
}
