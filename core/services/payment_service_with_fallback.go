package services

import (
	"context"
	"fmt"
	"time"

	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/repository"
)

// PaymentServiceWithFallback manages payment processing with fallback support
type PaymentServiceWithFallback struct {
	repo              repository.PaymentRepository
	defaultProcessor  domain.PaymentProcessor
	fallbackProcessor domain.PaymentProcessor
	circuitBreaker    *CircuitBreaker
	rateLimiter       *RateLimiter
}

// NewPaymentServiceWithFallback creates a new instance with fallback support
func NewPaymentServiceWithFallback(r repository.PaymentRepository, defaultProcessor domain.PaymentProcessor, fallbackProcessor domain.PaymentProcessor) *PaymentServiceWithFallback {
	return &PaymentServiceWithFallback{
		repo:              r,
		defaultProcessor:  defaultProcessor,
		fallbackProcessor: fallbackProcessor,
		circuitBreaker:    NewCircuitBreaker(3, 5*time.Second),
		rateLimiter:       NewRateLimiter(5),
	}
}

// Process processes a payment with fallback support
func (s *PaymentServiceWithFallback) Process(ctx context.Context, payment *domain.Payment) error {
	processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.rateLimiter.WithRateLimit(processCtx, func() error {
		return s.processPaymentWithFallback(processCtx, payment)
	})
}

// Summary returns payment summary (same as original PaymentService)
func (s *PaymentServiceWithFallback) Summary(ctx context.Context, from, to *time.Time) (*domain.PaymentSummary, error) {
	if from != nil && to != nil && from.After(*to) {
		return nil, fmt.Errorf("from date cannot be after to date")
	}
	return s.repo.Summary(ctx, from, to)
}

// processPaymentWithFallback tries default processor first, then fallback
func (s *PaymentServiceWithFallback) processPaymentWithFallback(ctx context.Context, payment *domain.Payment) error {
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
func (s *PaymentServiceWithFallback) tryProcessorWithCircuitBreaker(payment *domain.Payment, processor domain.PaymentProcessor) error {
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
