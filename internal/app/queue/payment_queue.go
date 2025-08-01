package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/services"
	"github.com/google/uuid"
)

type PaymentJob struct {
	ID      uuid.UUID
	Payment *domain.Payment
	Retries int
	Created time.Time
}

type PaymentQueue struct {
	jobs       chan PaymentJob
	workers    int
	service    *services.PaymentService
	stop       chan struct{}
	wg         sync.WaitGroup
	maxRetries int
	semaphore  chan struct{} // Semáforo para controlar concorrência de escrita no DB
}

func NewPaymentQueue(workers int, bufferSize int, service *services.PaymentService) *PaymentQueue {
	q := &PaymentQueue{
		jobs:       make(chan PaymentJob, bufferSize),
		workers:    workers,
		service:    service,
		stop:       make(chan struct{}),
		maxRetries: 3,
		semaphore:  make(chan struct{}, 2), // Máximo 2 escritas simultâneas no DB
	}

	for j := 0; j < workers; j++ {
		q.wg.Add(1)
		go q.worker(context.Background(), j)
	}

	return q
}

func (q *PaymentQueue) Enqueue(payment *domain.Payment) error {
	job := PaymentJob{
		ID:      uuid.New(),
		Payment: payment,
		Retries: 0,
		Created: time.Now(),
	}

	select {
	case q.jobs <- job:
		return nil
	default:
		return core.ErrQueueFull
	}
}

func (q *PaymentQueue) worker(ctx context.Context, workerID int) {
	defer q.wg.Done()

	for {
		select {
		case job := <-q.jobs:
			q.processJob(ctx, job, workerID)
		case <-q.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (q *PaymentQueue) processJob(ctx context.Context, job PaymentJob, workerID int) {
	// Controla quantas escritas simultâneas no DB
	q.semaphore <- struct{}{}
	defer func() { <-q.semaphore }()

	// Context com timeout para cada job
	jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Printf("[Worker %d] Processing job %s (attempt %d) - timestamp: %v", workerID, job.ID, job.Retries+1, time.Now().UnixNano())

	err := q.service.Process(jobCtx, job.Payment)
	if err != nil {
		log.Printf("[Worker %d] Failed to process payment for job %s: %v", workerID, job.ID, err)

		// Retry logic com backoff exponencial
		if job.Retries < q.maxRetries {
			job.Retries++

			// Backoff exponencial: 1s, 2s, 4s
			backoff := time.Duration(1<<job.Retries) * time.Second
			log.Printf("[Worker %d] Retrying job %s in %v", workerID, job.ID, backoff)

			go func() {
				time.Sleep(backoff)
				select {
				case q.jobs <- job:
				case <-q.stop:
				}
			}()
		} else {
			log.Printf("[Worker %d] Job %s failed after %d attempts, dropping", workerID, job.ID, q.maxRetries)
		}
		return
	}

	duration := time.Since(job.Created)
	log.Printf("[Worker %d] Successfully processed job %s in %v - timestamp: %v", workerID, job.ID, duration, time.Now().UnixNano())
}

func (q *PaymentQueue) Shutdown() {
	close(q.stop)
	q.wg.Wait()
}
