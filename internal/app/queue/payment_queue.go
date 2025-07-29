package queue

import (
	"context"
	"log"

	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/services"
	"github.com/google/uuid"
)

type PaymentJob struct {
	ID      uuid.UUID
	Payment *domain.Payment
}

type PaymentQueue struct {
	jobs    chan PaymentJob
	workers int
	service *services.PaymentService
	stop    chan struct{}
}

func NewPaymentQueue(workers int, bufferSize int, service *services.PaymentService) *PaymentQueue {
	q := &PaymentQueue{
		jobs:    make(chan PaymentJob, bufferSize),
		workers: workers,
		service: service,
		stop:    make(chan struct{}),
	}

	for j := 0; j < workers; j++ {
		go q.worker(context.Background())
	}

	return q
}

func (q *PaymentQueue) Enqueue(payment *domain.Payment) error {
	q.jobs <- PaymentJob{ID: uuid.New(), Payment: payment}

	return nil
}

func (q *PaymentQueue) worker(ctx context.Context) {
	for {
		select {
		case job := <-q.jobs:
			if err := q.service.Process(ctx, job.Payment); err != nil {
				log.Printf("Failed to process payment for job %v: %v", job.ID, err)
				continue
			}
		case <-q.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (q *PaymentQueue) Shutdown() {
	close(q.stop)
	close(q.jobs)
}
