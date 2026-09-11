package handlers

import (
	"net/http"
	"time"

	"findit-backend/config"
	"github.com/gin-gonic/gin"
)

// GetHealth checks API health status including DB connection ping
func GetHealth(c *gin.Context) {
	dbStatus := "connected"
	httpStatus := http.StatusOK
	statusText := "OK"

	if config.DB == nil {
		dbStatus = "disconnected"
		httpStatus = http.StatusServiceUnavailable
		statusText = "ERROR"
	} else {
		sqlDB, err := config.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "disconnected"
			httpStatus = http.StatusServiceUnavailable
			statusText = "ERROR"
		}
	}

	c.JSON(httpStatus, gin.H{
		"status":    statusText,
		"database":  dbStatus,
		"message":   "Find It API health status",
		"timestamp": time.Now(),
	})
}
