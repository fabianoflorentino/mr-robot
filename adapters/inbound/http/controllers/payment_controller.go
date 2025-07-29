package controllers

import (
	"net/http"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, correlationId and amount is required"})
		return
	}

	if err := u.q.Enqueue(payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to proccess payment:", "details": err.Error()})
	}

	c.Status(http.StatusAccepted)
}
