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

		// Legacy Items Endpoint
		api.GET("/items", handlers.GetItems)
		api.GET("/items/:id", handlers.GetItemByID)
		api.POST("/items", handlers.CreateItem)

		// Lost Items CRUD Endpoints
		api.GET("/lost-items", handlers.GetLostItems)
		api.GET("/lost-items/:id", handlers.GetLostItemByID)
		api.POST("/lost-items", handlers.CreateLostItem)
		api.PUT("/lost-items/:id", handlers.UpdateLostItem)
		api.DELETE("/lost-items/:id", handlers.DeleteLostItem)

		// Found Items CRUD Endpoints
		api.GET("/found-items", handlers.GetFoundItems)
		api.GET("/found-items/:id", handlers.GetFoundItemByID)
		api.POST("/found-items", handlers.CreateFoundItem)
		api.PUT("/found-items/:id", handlers.UpdateFoundItem)
		api.DELETE("/found-items/:id", handlers.DeleteFoundItem)
	}

	return r
}
