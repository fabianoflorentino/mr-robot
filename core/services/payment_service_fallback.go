package services

import (
	"context"
	"fmt"
	"time"

	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/repository"
)

// PaymentServiceFallback manages payment processing with fallback support
type PaymentServiceFallback struct {
	repo              repository.PaymentRepository
	defaultProcessor  domain.PaymentProcessor
	fallbackProcessor domain.PaymentProcessor
	circuitBreaker    *CircuitBreaker
	rateLimiter       *RateLimiter
}

// NewPaymentServiceFallback creates a new instance with fallback support
func NewPaymentServiceFallback(r repository.PaymentRepository, defaultProcessor domain.PaymentProcessor, fallbackProcessor domain.PaymentProcessor) *PaymentServiceFallback {
	return &PaymentServiceFallback{
		repo:              r,
		defaultProcessor:  defaultProcessor,
		fallbackProcessor: fallbackProcessor,
		circuitBreaker:    NewCircuitBreaker(3, 5*time.Second),
		rateLimiter:       NewRateLimiter(5),
	}
}

// Process processes a payment with fallback support
func (s *PaymentServiceFallback) Process(ctx context.Context, payment *domain.Payment) error {
	processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.rateLimiter.WithRateLimit(processCtx, func() error {
		return s.processPaymentWithFallback(processCtx, payment)
	})
}

// Summary returns payment summary (same as original PaymentService)
func (s *PaymentServiceFallback) Summary(ctx context.Context, from, to *time.Time) (*domain.PaymentSummary, error) {
	if from != nil && to != nil && from.After(*to) {
		return nil, fmt.Errorf("from date cannot be after to date")
	}
	return s.repo.Summary(ctx, from, to)
}

// processPaymentWithFallback tries default processor first, then fallback
func (s *PaymentServiceFallback) processPaymentWithFallback(ctx context.Context, payment *domain.Payment) error {
	// Try default processor first
	err := s.tryProcessorWithCircuitBreaker(payment, s.defaultProcessor)
	if err == nil {
		// Success with default processor
		return s.repo.Process(ctx, payment, s.defaultProcessor.ProcessorName())
	}

	// Default failed, try fallback processor
	fmt.Printf("Default processor failed: %v, trying fallback...\n", err)
	err = s.tryProcessorWithCircuitBreaker(payment, s.fallbackProcessor)
	if err == nil {
		// Success with fallback processor
		return s.repo.Process(ctx, payment, s.fallbackProcessor.ProcessorName())
	}

	// Both processors failed
	return fmt.Errorf("both default and fallback processors failed: %w", err)
}

// tryProcessorWithCircuitBreaker attempts to process with circuit breaker protection
func (s *PaymentServiceFallback) tryProcessorWithCircuitBreaker(payment *domain.Payment, processor domain.PaymentProcessor) error {
	return s.circuitBreaker.Call(func() error {
		ok, err := processor.Process(payment)
		if err != nil {
			return err
		}
		if !ok {
			return core.ErrPaymentProcessingFailed
		}
		return nil
	})
}
