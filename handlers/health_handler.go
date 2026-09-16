package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"findit-backend/config"
	"github.com/gin-gonic/gin"
)

// GetHealth checks API health status including DB connection ping.
// A short retry tolerates the pool briefly re-establishing a connection
// that MySQL closed while idle (stale pooled connection).
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
		if err == nil {
			const maxAttempts = 2
			const pingTimeout = 500 * time.Millisecond
			const retryDelay = 150 * time.Millisecond

			for attempt := 1; attempt <= maxAttempts; attempt++ {
				ctx, cancel := context.WithTimeout(c.Request.Context(), pingTimeout)
				pingErr := sqlDB.PingContext(ctx)
				cancel()

				if pingErr == nil {
					break
				}

				if attempt == maxAttempts {
					log.Printf("health: db ping gagal setelah %d percobaan: %v", attempt, pingErr)
					dbStatus = "disconnected"
					httpStatus = http.StatusServiceUnavailable
					statusText = "ERROR"
					break
				}

				time.Sleep(retryDelay)
			}
		} else {
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
