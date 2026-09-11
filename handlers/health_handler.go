package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetHealth checks API health status
func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"message":   "Find It API is running smoothly",
		"timestamp": time.Now(),
	})
}
