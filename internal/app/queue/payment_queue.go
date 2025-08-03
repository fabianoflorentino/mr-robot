package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/fabianoflorentino/mr-robot/config"
	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/internal/app/interfaces"
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
	service    interfaces.PaymentServiceInterface
	stop       chan struct{}
	wg         sync.WaitGroup
	maxRetries int
	semaphore  chan struct{}
	config     *config.QueueConfig
}

func NewPaymentQueue(queueConfig *config.QueueConfig, service interfaces.PaymentServiceInterface) *PaymentQueue {
	q := &PaymentQueue{
		jobs:       make(chan PaymentJob, queueConfig.BufferSize),
		workers:    queueConfig.Workers,
		service:    service,
		stop:       make(chan struct{}),
		maxRetries: queueConfig.MaxEnqueueRetries,
		semaphore:  make(chan struct{}, queueConfig.MaxSimultaneousWrites),
		config:     queueConfig,
	}

	for j := 0; j < queueConfig.Workers; j++ {
		q.wg.Add(1)
		go q.worker(context.Background(), j)
	}

	return q
}

func (q *PaymentQueue) Enqueue(payment *domain.Payment) error {
	job := PaymentJob{
		ID:      uuid.New(),
		Payment: payment,
		Retries: q.config.MaxEnqueueRetries,
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
	q.semaphore <- struct{}{}
	defer func() { <-q.semaphore }()

	jobCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Printf("[Worker %d] Processing job %s (attempt %d) - timestamp: %v", workerID, job.ID, job.Retries, time.Now().UnixNano())

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
