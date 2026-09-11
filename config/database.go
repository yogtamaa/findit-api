package config

import (
	"fmt"
	"log"
	"os"

	"findit-backend/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase initializes MySQL connection using GORM and performs auto-migration
func ConnectDatabase() *gorm.DB {
	// Load .env file if available
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ File .env tidak ditemukan, menggunakan environment variables sistem")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	if dbUser == "" {
		dbUser = "root"
	}
	if dbName == "" {
		dbName = "findit"
	}

	// Format Data Source Name (DSN)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Gagal terhubung ke database MySQL: %v", err)
	}

	log.Println("✅ Berhasil terhubung ke database MySQL")

	// Auto-migrate tables
	err = database.AutoMigrate(
		&models.LostItem{},
		&models.FoundItem{},
	)
	if err != nil {
		log.Printf("⚠️ Gagal melakukan auto migration: %v\n", err)
	} else {
		log.Println("✅ Auto migration tabel (lost_items & found_items) berhasil!")
	}

	DB = database
	return database
}
