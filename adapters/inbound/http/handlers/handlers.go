package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
)

func ShouldBindJSON(c *gin.Context, input any) error {
	if err := c.ShouldBindJSON(input); err != nil {
		log.Fatalf("error to parsed response: %v", err)
	}

	return nil
}
