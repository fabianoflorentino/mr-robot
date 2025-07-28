package services

import (
	"context"

	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/repository"
)

type PaymentService struct {
	repo      repository.PaymentRepository
	processor domain.PaymentProcessorDefault
}

func NewPaymentService(r repository.PaymentRepository, p domain.PaymentProcessorDefault) *PaymentService {
	return &PaymentService{repo: r, processor: p}
}

func (s *PaymentService) Process(ctx context.Context, payment *domain.Payment) error {
	ok, err := s.processor.Process(payment)
	if err != nil {
		return err
	}

	if !ok {
		return core.ErrPaymentProcessingFailed
	}

	if err := s.repo.Process(ctx, payment); err != nil {
		return err
	}

	return nil
}
