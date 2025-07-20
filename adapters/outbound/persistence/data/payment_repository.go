package data

import (
	"context"

	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/repository"
	"gorm.io/gorm"
)

type DataPaymentRepository struct {
	DB *gorm.DB
}

func NewDataPaymentRepository(db *gorm.DB) repository.PaymentRepository {
	return &DataPaymentRepository{DB: db}
}

func (d *DataPaymentRepository) Process(ctx context.Context, payment *domain.Payment) error {
	pymt := Payment{
		Amount: payment.Amount,
	}

	return d.DB.Create(&pymt).Error
}
