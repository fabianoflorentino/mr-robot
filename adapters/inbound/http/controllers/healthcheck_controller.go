package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthCheckController struct{}

func NewHealthCheckController() *HealthCheckController {
	return &HealthCheckController{}
}

func (h *HealthCheckController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": os.Getenv("HOSTNAME"),
		"status":  "ok",
		"time":    time.Now().Format(time.RFC3339),
	})
}
