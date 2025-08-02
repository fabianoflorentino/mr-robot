package repository

import (
	"context"

	"github.com/fabianoflorentino/mr-robot/core/domain"
)

type PaymentRepository interface {
	Process(ctx context.Context, payment *domain.Payment, processorName string) error
}
