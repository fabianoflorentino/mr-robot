package data

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CorrelationID uuid.UUID `json:"correlation_id" db:"correlation_id"`
	Amount        float64   `json:"amount" db:"amount"`
	Processor     string    `json:"processor" db:"processor"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
