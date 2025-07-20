package controllers

import (
	"net/http"

	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/services"
	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	srv *services.PaymentService
}

func NewPaymentController(p *services.PaymentService) *PaymentController {
	return &PaymentController{srv: p}
}

func (u *PaymentController) ProcessPayment(c *gin.Context) {
	var payment = &domain.Payment{}

	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request, amount is required"})
		return
	}

	_, err := u.srv.Process(c, payment)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}
