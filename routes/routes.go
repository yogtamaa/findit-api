package routes

import (
	"net/http"
	"os"
	"strings"

	"findit-backend/handlers"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware enables CORS with optional environment-configurable origins
func CORSMiddleware() gin.HandlerFunc {
	allowedOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv != "" {
		for _, o := range strings.Split(allowedOriginsEnv, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowOrigin := "*"

		if len(allowedOrigins) > 0 && origin != "" {
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowOrigin = origin
					break
				}
			}
		} else if origin != "" {
			allowOrigin = origin
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
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

// SecurityHeadersMiddleware adds basic production security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// SetupRouter initializes Gin router and endpoints
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Apply Middlewares
	r.Use(CORSMiddleware())
	r.Use(SecurityHeadersMiddleware())

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
