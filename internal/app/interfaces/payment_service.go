package interfaces

import (
	"context"
	"time"

	"github.com/fabianoflorentino/mr-robot/core/domain"
)

// PaymentServiceInterface defines the contract for payment services
type PaymentServiceInterface interface {
	Process(ctx context.Context, payment *domain.Payment) error
	Summary(ctx context.Context, from, to *time.Time) (*domain.PaymentSummary, error)
}
