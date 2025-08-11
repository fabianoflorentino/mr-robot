package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/repository"
	"github.com/google/uuid"
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
	DB *sql.DB
}

func NewDataPaymentRepository(db *sql.DB) repository.PaymentRepository {
	return &DataPaymentRepository{DB: db}
}

func (d *DataPaymentRepository) Process(ctx context.Context, payment *domain.Payment, processorName string) error {
	pymt := Payment{
		ID:            uuid.New(),
		CorrelationID: payment.CorrelationID,
		Amount:        payment.Amount,
		Processor:     processorName,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := d.retriesTransactions(ctx, &pymt); err != nil {
		return fmt.Errorf("failed to process payment: %w", err)
	}

	return nil
}

func (d *DataPaymentRepository) Summary(ctx context.Context, from, to *time.Time) (*domain.PaymentSummary, error) {
	var summary []struct {
		Processor     string  `db:"processor"`
		TotalAmount   float64 `db:"total_amount"`
		TotalRequests int64   `db:"total_requests"`
	}
	s := &domain.PaymentSummary{}

	query := `SELECT processor, SUM(amount) as total_amount, COUNT(*) as total_requests
	          FROM payments`

	var args []interface{}
	if from != nil && to != nil {
		query += ` WHERE created_at BETWEEN $1 AND $2`
		args = append(args, *from, *to)
	}

	query += ` GROUP BY processor`

	rows, err := d.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment summary: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r struct {
			Processor     string  `db:"processor"`
			TotalAmount   float64 `db:"total_amount"`
			TotalRequests int64   `db:"total_requests"`
		}

		if err := rows.Scan(&r.Processor, &r.TotalAmount, &r.TotalRequests); err != nil {
			return nil, fmt.Errorf("failed to scan payment summary row: %w", err)
		}

		summary = append(summary, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment summary rows: %w", err)
	}

	for _, r := range summary {
		switch r.Processor {
		case "default":
			s.Default = domain.ProcessorSummary{
				TotalRequests: r.TotalRequests,
				TotalAmount:   r.TotalAmount,
			}

		case "fallback":
			s.Fallback = domain.ProcessorSummary{
				TotalRequests: r.TotalRequests,
				TotalAmount:   r.TotalAmount,
			}
		default:
			return nil, fmt.Errorf("unknown processor: %s", r.Processor)
		}
	}

	return s, nil
}

func (d *DataPaymentRepository) Purge(ctx context.Context) error {
	query := `DELETE FROM payments`
	_, err := d.DB.ExecContext(ctx, query)
	return err
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
	tx, err := d.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Verifica se já existe (idempotência)
	var existingID uuid.UUID
	checkQuery := `SELECT id FROM payments WHERE correlation_id = $1 LIMIT 1`
	err = tx.QueryRowContext(ctx, checkQuery, pymt.CorrelationID).Scan(&existingID)

	if err == nil {
		// Já existe, não faz nada (idempotente)
		return tx.Commit()
	}

	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing payment: %w", err)
	}

	// Create a new payment record with a timeout context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	insertQuery := `INSERT INTO payments (id, correlation_id, amount, processor, created_at, updated_at) 
	                VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.ExecContext(ctxWithTimeout, insertQuery,
		pymt.ID, pymt.CorrelationID, pymt.Amount, pymt.Processor, pymt.CreatedAt, pymt.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	return tx.Commit()
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
