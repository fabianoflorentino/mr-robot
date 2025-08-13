package services

import (
	"database/sql"
	"fmt"

	"github.com/fabianoflorentino/mr-robot/adapters/outbound/gateway"
	"github.com/fabianoflorentino/mr-robot/adapters/outbound/persistence/data"
	"github.com/fabianoflorentino/mr-robot/core/services"
	"github.com/fabianoflorentino/mr-robot/internal/app/circuitbreaker"
	"github.com/fabianoflorentino/mr-robot/internal/app/interfaces"
	"github.com/fabianoflorentino/mr-robot/internal/app/payment"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
)

// Manager handles service initialization and management
type Manager struct {
	db                   *sql.DB
	paymentConfig        *payment.Config
	queueConfig          *queue.Config
	circuitBreakerConfig *circuitbreaker.Config
	paymentService       interfaces.PaymentServiceInterface
	paymentQueue         *queue.PaymentQueue
}

// NewManager creates a new service manager
func NewManager(db *sql.DB, paymentConfig *payment.Config, queueConfig *queue.Config, circuitBreakerConfig *circuitbreaker.Config) *Manager {
	return &Manager{
		db:                   db,
		paymentConfig:        paymentConfig,
		queueConfig:          queueConfig,
		circuitBreakerConfig: circuitBreakerConfig,
	}
}

// InitializeServices sets up all application services
func (s *Manager) InitializeServices() error {
	// Initialize payment service first (queue depends on it)
	if err := s.initializePaymentService(); err != nil {
		return fmt.Errorf("failed to initialize payment service: %w", err)
	}

	// Initialize payment queue
	if err := s.initializePaymentQueue(); err != nil {
		return fmt.Errorf("failed to initialize payment queue: %w", err)
	}

	return nil
}

// initializePaymentService creates and configures the payment service with fallback
func (s *Manager) initializePaymentService() error {
	paymentRepo := data.NewDataPaymentRepository(s.db)

	// Create default processor
	defaultProcessor := &gateway.ProcessGateway{
		URL:  s.paymentConfig.DefaultProcessorURL,
		Name: "default",
	}

	// Create fallback processor
	fallbackProcessor := &gateway.ProcessGateway{
		URL:  s.paymentConfig.FallbackProcessorURL,
		Name: "fallback",
	}

	// Convert circuit breaker config to legacy format
	// Use the new service with fallback support
	s.paymentService = services.NewPaymentService(paymentRepo, defaultProcessor, fallbackProcessor, s.circuitBreakerConfig)

	return nil
}

// initializePaymentQueue creates and configures the payment queue
func (s *Manager) initializePaymentQueue() error {
	s.paymentQueue = queue.NewPaymentQueue(s.queueConfig, s.paymentService)

	return nil
}

// GetPaymentService returns the payment service instance
func (s *Manager) GetPaymentService() interfaces.PaymentServiceInterface {
	return s.paymentService
}

// GetPaymentQueue returns the payment queue instance
func (s *Manager) GetPaymentQueue() *queue.PaymentQueue {
	return s.paymentQueue
}

// Shutdown gracefully shuts down all services
func (s *Manager) Shutdown() {
	if s.paymentQueue != nil {
		s.paymentQueue.Shutdown()
	}
}
