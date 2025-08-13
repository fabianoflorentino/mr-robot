package config

import (
	"fmt"

	"github.com/fabianoflorentino/mr-robot/internal/app/circuitbreaker"
	"github.com/fabianoflorentino/mr-robot/internal/app/controller"
	"github.com/fabianoflorentino/mr-robot/internal/app/database"
	"github.com/fabianoflorentino/mr-robot/internal/app/payment"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
)

// Manager handles centralized configuration loading and management
type Manager struct {
	databaseManager       *database.ConfigManager
	paymentManager        *payment.ConfigManager
	queueManager          *queue.ConfigManager
	circuitBreakerManager *circuitbreaker.ConfigManager
	controllerManager     *controller.ConfigManager
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		databaseManager:       database.NewConfigManager(),
		paymentManager:        payment.NewConfigManager(),
		queueManager:          queue.NewConfigManager(),
		circuitBreakerManager: circuitbreaker.NewConfigManager(),
		controllerManager:     controller.NewConfigManager(),
	}
}

// LoadConfiguration loads all application configurations
func (m *Manager) LoadConfiguration() error {
	// Load database configuration
	if err := m.databaseManager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load database configuration: %w", err)
	}

	// Load payment configuration
	if err := m.paymentManager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load payment configuration: %w", err)
	}

	// Load queue configuration
	if err := m.queueManager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load queue configuration: %w", err)
	}

	// Load circuit breaker configuration
	if err := m.circuitBreakerManager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load circuit breaker configuration: %w", err)
	}

	// Load controller configuration
	if err := m.controllerManager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load controller configuration: %w", err)
	}

	return nil
}

// ValidateConfiguration validates all configurations
func (m *Manager) ValidateConfiguration() error {
	if err := m.databaseManager.Validate(); err != nil {
		return fmt.Errorf("invalid database configuration: %w", err)
	}

	if err := m.paymentManager.Validate(); err != nil {
		return fmt.Errorf("invalid payment configuration: %w", err)
	}

	if err := m.queueManager.Validate(); err != nil {
		return fmt.Errorf("invalid queue configuration: %w", err)
	}

	if err := m.circuitBreakerManager.Validate(); err != nil {
		return fmt.Errorf("invalid circuit breaker configuration: %w", err)
	}

	if err := m.controllerManager.Validate(); err != nil {
		return fmt.Errorf("invalid controller configuration: %w", err)
	}

	return nil
}

// GetDatabaseConfig returns the database configuration
func (m *Manager) GetDatabaseConfig() *database.Config {
	return m.databaseManager.GetConfig()
}

// GetPaymentConfig returns the payment configuration
func (m *Manager) GetPaymentConfig() *payment.Config {
	return m.paymentManager.GetConfig()
}

// GetQueueConfig returns the queue configuration
func (m *Manager) GetQueueConfig() *queue.Config {
	return m.queueManager.GetConfig()
}

// GetCircuitBreakerConfig returns the circuit breaker configuration
func (m *Manager) GetCircuitBreakerConfig() *circuitbreaker.Config {
	return m.circuitBreakerManager.GetConfig()
}

// GetControllerConfig returns the controller configuration
func (m *Manager) GetControllerConfig() *controller.Config {
	return m.controllerManager.GetConfig()
}

// GetDatabaseManager returns the database config manager
func (m *Manager) GetDatabaseManager() *database.ConfigManager {
	return m.databaseManager
}

// GetPaymentManager returns the payment config manager
func (m *Manager) GetPaymentManager() *payment.ConfigManager {
	return m.paymentManager
}

// GetQueueManager returns the queue config manager
func (m *Manager) GetQueueManager() *queue.ConfigManager {
	return m.queueManager
}

// GetCircuitBreakerManager returns the circuit breaker config manager
func (m *Manager) GetCircuitBreakerManager() *circuitbreaker.ConfigManager {
	return m.circuitBreakerManager
}

// GetControllerManager returns the controller config manager
func (m *Manager) GetControllerManager() *controller.ConfigManager {
	return m.controllerManager
}
