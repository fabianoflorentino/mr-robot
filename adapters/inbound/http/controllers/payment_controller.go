package controllers

import (
	"net/http"
	"time"

	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	q *queue.PaymentQueue
}

func NewPaymentController(q *queue.PaymentQueue) *PaymentController {
	return &PaymentController{q: q}
}

func (u *PaymentController) ProcessPayment(c *gin.Context) {
	var payment = &domain.Payment{}

	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "correlationId and amount are required"})
		return
	}

	u.enqueuePaymentWithTimeout(c, payment)
}

func (u *PaymentController) enqueuePaymentWithTimeout(c *gin.Context, payment *domain.Payment) {
	eq := make(chan error, 1)

	go func() { eq <- u.q.Enqueue(payment) }()

	select {
	case err := <-eq:
		if err != nil {
			if err == core.ErrQueueFull {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "system is busy, please try again later"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment", "details": err.Error()})
			return
		}

		c.Status(http.StatusAccepted)

	case <-time.After(5 * time.Second):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "request timeout", "details": "unable to queue payment within timeout"})
	}
}
