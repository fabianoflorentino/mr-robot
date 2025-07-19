package data

import "github.com/google/uuid"

type Payment struct {
	CorrelationID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Amount        float64   `gorm:"not null"`
}
