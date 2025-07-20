package data

import "github.com/google/uuid"

type Payment struct {
	ID            uuid.UUID `gorm:"primaryKey,type:uuid;default:gen_random_uuid()"`
	CorrelationID uuid.UUID `gorm:"not null"`
	Amount        float64   `gorm:"not null"`
}
