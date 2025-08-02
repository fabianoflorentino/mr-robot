package services

import (
	"fmt"

	"github.com/fabianoflorentino/mr-robot/adapters/outbound/gateway"
	"github.com/fabianoflorentino/mr-robot/adapters/outbound/persistence/data"
	"github.com/fabianoflorentino/mr-robot/config"
	"github.com/fabianoflorentino/mr-robot/core/services"
	"github.com/fabianoflorentino/mr-robot/internal/app/interfaces"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
	"gorm.io/gorm"
)

// Manager handles service initialization and management
type Manager struct {
	config         *config.AppConfig
	db             *gorm.DB
	paymentService interfaces.PaymentServiceInterface
	paymentQueue   *queue.PaymentQueue
}

// NewManager creates a new service manager
func NewManager(cfg *config.AppConfig, db *gorm.DB) *Manager {
	return &Manager{
		config: cfg,
		db:     db,
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
		URL:  s.config.Payment.DefaultProcessorURL,
		Name: "default",
	}

	// Create fallback processor
	fallbackProcessor := &gateway.ProcessGateway{
		URL:  s.config.Payment.FallbackProcessorURL,
		Name: "fallback",
	}

	// Use the new service with fallback support
	s.paymentService = services.NewPaymentServiceWithFallback(paymentRepo, defaultProcessor, fallbackProcessor)

	return nil
}

// initializePaymentQueue creates and configures the payment queue
func (s *Manager) initializePaymentQueue() error {
	s.paymentQueue = queue.NewPaymentQueue(s.config.Queue.Workers, s.config.Queue.BufferSize, s.paymentService)

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
