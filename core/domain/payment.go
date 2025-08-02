package domain

import "github.com/google/uuid"

type Payment struct {
	CorrelationID uuid.UUID `json:"correlationId" binding:"required"`
	Amount        float64   `json:"amount" binding:"required,gt=0"`
}

type PaymentSummary struct {
	Default  ProcessorSummary `json:"default"`
	Fallback ProcessorSummary `json:"fallback"`
}

type ProcessorSummary struct {
	TotalRequests int64   `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentProcessor interface {
	Process(payment *Payment) (bool, error)
	ProcessorName() string
}
