package domain

import "github.com/google/uuid"

type Payment struct {
	CorrelationID uuid.UUID
	amount        float64
}
