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
	repo                   repository.PaymentRepository
	defaultProcessor       domain.PaymentProcessor
	fallbackProcessor      domain.PaymentProcessor
	defaultCircuitBreaker  *CircuitBreaker
	fallbackCircuitBreaker *CircuitBreaker
	rateLimiter            *RateLimiter
}

// NewPaymentServiceFallback creates a new instance with fallback support
func NewPaymentServiceFallback(r repository.PaymentRepository, defaultProcessor domain.PaymentProcessor, fallbackProcessor domain.PaymentProcessor) *PaymentServiceFallback {
	return &PaymentServiceFallback{
		repo:                   r,
		defaultProcessor:       defaultProcessor,
		fallbackProcessor:      fallbackProcessor,
		defaultCircuitBreaker:  NewCircuitBreaker(3, 3*time.Second), // Faster reset for default
		fallbackCircuitBreaker: NewCircuitBreaker(3, 3*time.Second), // Faster reset for fallback
		rateLimiter:            NewRateLimiter(10),                  // Increased concurrency
	}
}

// Process processes a payment with fallback support
func (s *PaymentServiceFallback) Process(ctx context.Context, payment *domain.Payment) error {
	processCtx, cancel := context.WithTimeout(ctx, 3*time.Second) // Reduced timeout for faster processing
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
	// Try default processor first with its own circuit breaker
	err := s.tryProcessorWithCircuitBreaker(payment, s.defaultProcessor, s.defaultCircuitBreaker)
	if err == nil {
		// Success with default processor
		return s.repo.Process(ctx, payment, s.defaultProcessor.ProcessorName())
	}

	// Default failed, try fallback processor with its own circuit breaker
	fmt.Printf("Default processor failed: %v, trying fallback...\n", err)
	err = s.tryProcessorWithCircuitBreaker(payment, s.fallbackProcessor, s.fallbackCircuitBreaker)
	if err == nil {
		// Success with fallback processor
		return s.repo.Process(ctx, payment, s.fallbackProcessor.ProcessorName())
	}

	// Both processors failed
	return fmt.Errorf("both default and fallback processors failed: %w", err)
}

// tryProcessorWithCircuitBreaker attempts to process with circuit breaker protection
func (s *PaymentServiceFallback) tryProcessorWithCircuitBreaker(payment *domain.Payment, processor domain.PaymentProcessor, circuitBreaker *CircuitBreaker) error {
	return circuitBreaker.Call(func() error {
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

// GetDefaultCircuitBreakerState returns the state of the default processor circuit breaker
func (s *PaymentServiceFallback) GetDefaultCircuitBreakerState() CircuitBreakerState {
	return s.defaultCircuitBreaker.GetState()
}

// GetFallbackCircuitBreakerState returns the state of the fallback processor circuit breaker
func (s *PaymentServiceFallback) GetFallbackCircuitBreakerState() CircuitBreakerState {
	return s.fallbackCircuitBreaker.GetState()
}

// GetDefaultCircuitBreakerFailureCount returns the failure count of the default processor circuit breaker
func (s *PaymentServiceFallback) GetDefaultCircuitBreakerFailureCount() int {
	return s.defaultCircuitBreaker.GetFailureCount()
}

// GetFallbackCircuitBreakerFailureCount returns the failure count of the fallback processor circuit breaker
func (s *PaymentServiceFallback) GetFallbackCircuitBreakerFailureCount() int {
	return s.fallbackCircuitBreaker.GetFailureCount()
}
