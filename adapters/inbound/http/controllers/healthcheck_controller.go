package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func (u *PaymentController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": os.Getenv("HOSTNAME"),
		"status":  "ok",
		"time":    time.Now().Format(time.RFC3339),
	})
}
