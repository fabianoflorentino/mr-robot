package services

import (
	"fmt"

	"github.com/fabianoflorentino/mr-robot/adapters/outbound/gateway"
	"github.com/fabianoflorentino/mr-robot/adapters/outbound/persistence/data"
	"github.com/fabianoflorentino/mr-robot/config"
	"github.com/fabianoflorentino/mr-robot/core/services"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
	"gorm.io/gorm"
)

// Manager handles service initialization and management
type Manager struct {
	config         *config.AppConfig
	db             *gorm.DB
	paymentService *services.PaymentService
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

// initializePaymentService creates and configures the payment service
func (s *Manager) initializePaymentService() error {
	paymentRepo := data.NewDataPaymentRepository(s.db)
	processor := gateway.NewDefaultProcessor(s.config.Payment.DefaultProcessorURL)
	s.paymentService = services.NewPaymentService(paymentRepo, &processor)

	return nil
}

// initializePaymentQueue creates and configures the payment queue
func (s *Manager) initializePaymentQueue() error {
	s.paymentQueue = queue.NewPaymentQueue(s.config.Queue.Workers, s.config.Queue.BufferSize, s.paymentService)

	return nil
}

// GetPaymentService returns the payment service instance
func (s *Manager) GetPaymentService() *services.PaymentService {
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
