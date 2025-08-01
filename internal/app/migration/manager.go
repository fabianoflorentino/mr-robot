package migration

import (
	"fmt"
	"log"

	"github.com/fabianoflorentino/mr-robot/adapters/outbound/persistence/data"
	"gorm.io/gorm"
)

// Manager handles database migrations
type Manager struct {
	db *gorm.DB
}

// NewManager creates a new migration manager
func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db: db,
	}
}

// RunMigrations executes database migrations
func (m *Manager) RunMigrations() error {
	if err := m.db.AutoMigrate(&data.Payment{}); err != nil {
		return fmt.Errorf("failed to migrate payment model: %w", err)
	} else {
		log.Println("Database migrations completed successfully")
	}

	return nil
}
