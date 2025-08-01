package services

import (
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the circuit breaker states
type CircuitBreakerState int

const (
	Closed CircuitBreakerState = iota
	Open
	HalfOpen
)

// CircuitBreaker implements the Circuit Breaker pattern for fast failures
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failureCount int
	lastFailTime time.Time
	state        CircuitBreakerState
	mutex        sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker instance
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        Closed,
	}
}

// Call executes a function protected by the circuit breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Check if we should reset the circuit breaker
	if cb.state == Open && time.Since(cb.lastFailTime) > cb.resetTimeout {
		cb.state = HalfOpen
		cb.failureCount = 0
	}

	// If circuit is open, fail fast
	if cb.state == Open {
		return fmt.Errorf("circuit breaker is open")
	}

	// Execute the function
	err := fn()
	if err != nil {
		cb.failureCount++
		cb.lastFailTime = time.Now()

		if cb.failureCount >= cb.maxFailures {
			cb.state = Open
		}
		return err
	}

	// Success - reset if we were in half-open state
	if cb.state == HalfOpen {
		cb.state = Closed
	}
	cb.failureCount = 0
	return nil
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetFailureCount returns the current number of failures
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failureCount
}
