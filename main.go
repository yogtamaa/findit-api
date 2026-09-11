package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"findit-backend/config"
	"findit-backend/routes"
)

func main() {
	// Initialize database connection and auto migrations
	config.ConnectDatabase()

	// Port configuration via environment variable or default 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := routes.SetupRouter()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Initializing the server in a goroutine so that it doesn't block graceful shutdown
	go func() {
		fmt.Printf("🚀 Server Find It Backend berjalan pada port :%s...\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Gagal menjalankan server: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⚠️ Shutting down server gracefully...")

	// 10 second timeout context for pending requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v\n", err)
	}

	// Close database connection gracefully
	if config.DB != nil {
		if sqlDB, err := config.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("⚠️ Error closing database connection: %v\n", err)
			} else {
				log.Println("✅ Database connection closed cleanly")
			}
		}
	}

	log.Println("✅ Server exited cleanly")
}
