package data

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID            uuid.UUID `gorm:"primaryKey,type:uuid;default:gen_random_uuid()"`
	CorrelationID uuid.UUID `gorm:"not null"`
	Amount        float64   `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}
