package services

import (
	"context"
	"fmt"
	"time"

	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/repository"
)

// PaymentService manages payment processing
type PaymentService struct {
	repo           repository.PaymentRepository
	processor      domain.PaymentProcessor
	processorName  string
	circuitBreaker *CircuitBreaker
	rateLimiter    *RateLimiter
}

// NewPaymentService creates a new instance of the payment service
func NewPaymentService(r repository.PaymentRepository, p domain.PaymentProcessor) *PaymentService {
	return &PaymentService{
		repo:           r,
		processor:      p,
		processorName:  p.ProcessorName(),
		circuitBreaker: NewCircuitBreaker(3, 3*time.Second), // Optimized: 3 failures in 3 seconds
		rateLimiter:    NewRateLimiter(10),                  // Increased: Max 10 concurrent processings
	}
}

// Process processes a payment with circuit breaker and rate limiting protections
func (s *PaymentService) Process(ctx context.Context, payment *domain.Payment) error {
	// Context with timeout for the entire processing
	processCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Rate limiting - limits concurrent processings
	return s.rateLimiter.WithRateLimit(processCtx, func() error {
		return s.processPayment(processCtx, payment)
	})
}

func (s *PaymentService) Summary(ctx context.Context, from, to *time.Time) (*domain.PaymentSummary, error) {
	// Validate time range
	if from != nil && to != nil && from.After(*to) {
		return nil, fmt.Errorf("from date cannot be after to date")
	}

	return s.repo.Summary(ctx, from, to)
}

// processPayment executes the payment processing
func (s *PaymentService) processPayment(ctx context.Context, payment *domain.Payment) error {
	// Circuit breaker for external processing
	if err := s.circuitBreaker.Call(func() error { return s.processWithExternalService(payment) }); err != nil {
		return fmt.Errorf("payment processing failed: %w", err)
	}

	// Save to database with automatic retry (implemented in repository)
	if err := s.repo.Process(ctx, payment, s.processorName); err != nil {
		return fmt.Errorf("failed to persist payment: %w", err)
	}

	return nil
}

// processWithExternalService processes the payment with external service
func (s *PaymentService) processWithExternalService(payment *domain.Payment) error {
	ok, err := s.processor.Process(payment)
	if err != nil {
		return err
	}
	if !ok {
		return core.ErrPaymentProcessingFailed
	}
	return nil
}
