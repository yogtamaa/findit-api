package main

import (
	"fmt"
	"log"
	"os"

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

	fmt.Printf("🚀 Server Find It Backend berjalan pada port :%s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
