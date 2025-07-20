package services

import (
	"context"

	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/repository"
)

type PaymentService struct {
	payment repository.PaymentRepository
}

func NewPaymentService(p repository.PaymentRepository) *PaymentService {
	return &PaymentService{payment: p}
}

func (p *PaymentService) Process(ctx context.Context, payment *domain.Payment) (*domain.Payment, error) {
	if err := p.payment.Process(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}
