package services

import (
	"context"
)

// RateLimiter implements rate control for concurrent processing
type RateLimiter struct {
	limiter chan struct{}
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(maxConcurrency int) *RateLimiter {
	return &RateLimiter{
		limiter: make(chan struct{}, maxConcurrency),
	}
}

// Acquire attempts to acquire a slot for processing
func (rl *RateLimiter) Acquire(ctx context.Context) error {
	select {
	case rl.limiter <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release releases a processing slot
func (rl *RateLimiter) Release() {
	<-rl.limiter
}

// WithRateLimit executes a function with rate limiting
func (rl *RateLimiter) WithRateLimit(ctx context.Context, fn func() error) error {
	if err := rl.Acquire(ctx); err != nil {
		return err
	}
	defer rl.Release()

	return fn()
}
