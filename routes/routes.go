package routes

import (
	"net/http"

	"findit-backend/handlers"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware enables CORS for frontend requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SetupRouter initializes Gin router and endpoints
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Apply CORS Middleware
	r.Use(CORSMiddleware())

	// API Routes Group
	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", handlers.GetHealth)

		// Auth
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)

		// Users
		api.GET("/users/:id", handlers.GetUserByID)

		// Reports (single table for both lost & found, filter with ?type=lost|found)
		api.GET("/reports", handlers.GetReports)
		api.GET("/reports/:id", handlers.GetReportByID)
		api.POST("/reports", handlers.CreateReport)
		api.PUT("/reports/:id", handlers.UpdateReport)
		api.DELETE("/reports/:id", handlers.DeleteReport)

		// Categories
		api.GET("/categories", handlers.GetCategories)

		// Matches
		api.GET("/matches", handlers.GetMatches)
		api.GET("/matches/:id", handlers.GetMatchByID)
		api.POST("/matches", handlers.CreateMatch)
		api.PUT("/matches/:id", handlers.UpdateMatch)

		// Notifications
		api.GET("/notifications/user/:userId", handlers.GetNotificationsForUser)
		api.POST("/notifications", handlers.CreateNotification)
		api.PUT("/notifications/:id/read", handlers.MarkNotificationRead)
	}

	return r
}
