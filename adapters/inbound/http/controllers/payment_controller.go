package controllers

import (
	"net/http"
	"time"

	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/internal/app/interfaces"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	q *queue.PaymentQueue
	s interfaces.PaymentServiceInterface
}

func NewPaymentController(q *queue.PaymentQueue, s interfaces.PaymentServiceInterface) *PaymentController {
	return &PaymentController{q: q, s: s}
}

func (u *PaymentController) PaymentProcess(c *gin.Context) {
	var payment = &domain.Payment{}

	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "correlationId and amount are required"})
		return
	}

	u.enqueuePaymentWithTimeout(c, payment)
}

func (u *PaymentController) PaymentsSummary(c *gin.Context) {
	var from, to *time.Time

	queryFrom := c.Query("from")
	queryTo := c.Query("to")

	// Parse query parameters for date range
	// If both are provided, parse them and set the from and to variables
	if queryFrom != "" && queryTo != "" {
		fromParsed, err := time.Parse(time.RFC3339, queryFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format, use RFC3339 format"})
			return
		}

		toParsed, err := time.Parse(time.RFC3339, queryTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format, use RFC3339 format"})
			return
		}

		from = &fromParsed
		to = &toParsed
	} else if queryFrom != "" || queryTo != "" {
		// If only one of the dates is provided, return an error
		c.JSON(http.StatusBadRequest, gin.H{"error": "both from and to dates must be provided"})
		return
	}

	summary, err := u.s.Summary(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve payment summary", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
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
