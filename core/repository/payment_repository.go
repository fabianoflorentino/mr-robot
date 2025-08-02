package repository

import (
	"context"
	"time"

	"github.com/fabianoflorentino/mr-robot/core/domain"
)

type PaymentRepository interface {
	Process(ctx context.Context, payment *domain.Payment, processorName string) error
	Summary(ctx context.Context, from, to *time.Time) (*domain.PaymentSummary, error)
}
