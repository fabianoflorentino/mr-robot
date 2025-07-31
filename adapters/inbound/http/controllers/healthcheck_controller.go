package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (u *PaymentController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "mr_robot",
		"status":  "ok",
		"time":    time.Now().Format(time.RFC3339),
	})
}
