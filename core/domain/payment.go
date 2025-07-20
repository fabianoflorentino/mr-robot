package domain

import "github.com/google/uuid"

type Payment struct {
	CorrelationID uuid.UUID `json:"correlation_id" binding:"required"`
	Amount        float64   `json:"amount" binding:"required,gt=0"`
}
