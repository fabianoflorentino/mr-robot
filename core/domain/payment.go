package domain

import "github.com/google/uuid"

type Payment struct {
	CorrelationID uuid.UUID `json:"correlationId" binding:"required"`
	Amount        float64   `json:"amount" binding:"required,gt=0"`
}

type PaymentProcessor interface {
	Process(payment *Payment) (bool, error)
	ProcessorName() string
}
