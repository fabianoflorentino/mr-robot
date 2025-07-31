package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/repository"
	"gorm.io/gorm"
)

var (
	maxRetries            int      = 3
	deadlockErrorPatterns []string = []string{
		"deadlock detected",
		"could not serialize access",
		"concurrent update",
	}
)

type DataPaymentRepository struct {
	DB *gorm.DB
}

func NewDataPaymentRepository(db *gorm.DB) repository.PaymentRepository {
	return &DataPaymentRepository{DB: db}
}

func (d *DataPaymentRepository) Process(ctx context.Context, payment *domain.Payment) error {
	pymt := Payment{
		CorrelationID: payment.CorrelationID,
		Amount:        payment.Amount,
	}

	if err := d.retriesTransactions(ctx, &pymt); err != nil {
		return fmt.Errorf("failed to process payment: %w", err)
	}

	return nil
}

// retriesTransactions try to process the payment with retries in case of deadlocks
// It uses exponential backoff for retries
func (d *DataPaymentRepository) retriesTransactions(ctx context.Context, pymt *Payment) error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := d.processWithTransaction(ctx, pymt)
		if err == nil {
			return nil
		}

		// if the error is a deadlock, we retry with exponential backoff
		if d.isDeadlockError(err) && attempt < maxRetries {
			// Backoff exponencial: 100ms, 200ms, 400ms
			backoff := time.Duration(100*attempt*attempt) * time.Millisecond
			time.Sleep(backoff)
			continue
		}

		return fmt.Errorf("failed to process payment after %d attempts: %w", attempt, err)
	}

	return nil
}

// processWithTransaction processes the payment within a transaction
// It checks for idempotency by looking for existing records with the same CorrelationID
func (d *DataPaymentRepository) processWithTransaction(ctx context.Context, pymt *Payment) error {
	return d.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verifica se já existe (idempotência)
		var existing Payment
		err := tx.Where("correlation_id = ?", pymt.CorrelationID).First(&existing).Error
		if err == nil {
			// Já existe, não faz nada (idempotente)
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		// Create a new payment record with a timeout context
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		return tx.WithContext(ctxWithTimeout).Create(pymt).Error
	}, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
}

// isDeadlockError checks if the error is a deadlock or serialization error
func (d *DataPaymentRepository) isDeadlockError(err error) bool {
	if err == nil {
		return false
	}

	return d.containsAnyErrorPattern(err.Error(), deadlockErrorPatterns)
}

// containsAnyErrorPattern checks if the error message contains any of the specified patterns
func (d *DataPaymentRepository) containsAnyErrorPattern(errMsg string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}
	return false
}
