package gateway

import (
	"time"
)

// ProcessorType represents the type of processor
type ProcessorType string

const (
	DefaultProcessor  ProcessorType = "default"
	FallbackProcessor ProcessorType = "fallback"
)

// ProcessorConfig holds configuration for a processor
type ProcessorConfig struct {
	URL     string
	Timeout time.Duration
}

// ProcessorFactory creates processors based on type and configuration
type ProcessorFactory struct{}

// NewProcessorFactory creates a new processor factory
func NewProcessorFactory() *ProcessorFactory {
	return &ProcessorFactory{}
}

// CreateProcessor creates a processor of the specified type with the given configuration
func (f *ProcessorFactory) CreateProcessor(processorType ProcessorType, config ProcessorConfig) *ProcessGateway {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return &ProcessGateway{
		URL:     config.URL,
		Name:    string(processorType),
		timeout: config.Timeout,
	}
}
