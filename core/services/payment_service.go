package services

import (
	"context"
	"fmt"
	"time"

	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/repository"
	"github.com/fabianoflorentino/mr-robot/internal/app/circuitbreaker"
)

// PaymentService manages payment processing with fallback support
type PaymentService struct {
	repo                   repository.PaymentRepository
	defaultProcessor       domain.PaymentProcessor
	fallbackProcessor      domain.PaymentProcessor
	defaultCircuitBreaker  *CircuitBreaker
	fallbackCircuitBreaker *CircuitBreaker
	rateLimiter            *RateLimiter
	config                 *circuitbreaker.Config
}

// NewPaymentService creates a new instance with fallback support
func NewPaymentService(
	r repository.PaymentRepository,
	defaultProcessor domain.PaymentProcessor,
	fallbackProcessor domain.PaymentProcessor,
	cfg *circuitbreaker.Config,
) *PaymentService {

	return &PaymentService{
		repo:                   r,
		defaultProcessor:       defaultProcessor,
		fallbackProcessor:      fallbackProcessor,
		defaultCircuitBreaker:  NewCircuitBreaker(cfg.MaxFailures, cfg.ResetTimeout),
		fallbackCircuitBreaker: NewCircuitBreaker(cfg.MaxFailures, cfg.ResetTimeout),
		rateLimiter:            NewRateLimiter(cfg.RateLimit),
		config:                 cfg,
	}
}

// Process processes a payment with fallback support
func (s *PaymentService) Process(ctx context.Context, payment *domain.Payment) error {
	processCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	return s.rateLimiter.WithRateLimit(processCtx, func() error {
		return s.processPayment(processCtx, payment)
	})
}

// Summary returns payment summary (same as original PaymentService)
func (s *PaymentService) Summary(ctx context.Context, from, to *time.Time) (*domain.PaymentSummary, error) {
	if from != nil && to != nil && from.After(*to) {
		return nil, fmt.Errorf("from date cannot be after to date")
	}
	return s.repo.Summary(ctx, from, to)
}

func (s *PaymentService) Purge(ctx context.Context) error {
	return s.repo.Purge(ctx)
}

// processPayment tries default processor first, then fallback
func (s *PaymentService) processPayment(ctx context.Context, payment *domain.Payment) error {
	// Try default processor first with its own circuit breaker
	err := s.tryProcessorWithCircuitBreaker(payment, s.defaultProcessor, s.defaultCircuitBreaker)
	if err == nil {
		// Success with default processor
		return s.repo.Process(ctx, payment, s.defaultProcessor.ProcessorName())
	}

	// Default failed, try fallback processor with its own circuit breaker
	fmt.Printf("Default processor (%s) failed: %v, trying fallback...\n", s.defaultProcessor.ProcessorName(), err)
	if err = s.tryProcessorWithCircuitBreaker(payment, s.fallbackProcessor, s.fallbackCircuitBreaker); err == nil {
		// Success with fallback processor
		return s.repo.Process(ctx, payment, s.fallbackProcessor.ProcessorName())
	}

	// Both processors failed
	return fmt.Errorf("both default and fallback processors failed: %w", err)
}

// tryProcessorWithCircuitBreaker attempts to process with circuit breaker protection
func (s *PaymentService) tryProcessorWithCircuitBreaker(payment *domain.Payment, processor domain.PaymentProcessor, circuitBreaker *CircuitBreaker) error {
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
