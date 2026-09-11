package handlers

import "net/http"

import (
	"fmt"
	"sync"
	"time"

	"findit-backend/models"
	"github.com/gin-gonic/gin"
)

var (
	itemsMutex sync.RWMutex
	items      = []models.Item{
		{
			ID:          "1",
			Title:       "Dompet Kulit Hitam",
			Description: "Tertinggal dompet kulit merek Eiger berisi KTP dan kartu ATM.",
			Category:    "wallet",
			Status:      "LOST",
			Location:    "Kantin Lt. 2",
			Contact:     "081234567890",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          "2",
			Title:       "Kunci Motor Honda",
			Description: "Ditemukan kunci motor Honda dengan gantungan akrilik anime.",
			Category:    "keys",
			Status:      "FOUND",
			Location:    "Parkiran Motor Barat",
			Contact:     "089876543210",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
		},
		{
			ID:          "3",
			Title:       "Earphone Wireless TWS",
			Description: "Hilang case earphone warna putih merek Anker di perpustakaan.",
			Category:    "electronics",
			Status:      "LOST",
			Location:    "Perpustakaan Lantai 1",
			Contact:     "085678901234",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		},
	}
	nextID = 4
)

// GetHealth checks API health status
func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"message":   "Find It API is running smoothly",
		"timestamp": time.Now(),
	})
}

// GetItems retrieves all lost and found item reports
func GetItems(c *gin.Context) {
	itemsMutex.RLock()
	defer itemsMutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})
}

// GetItemByID retrieves a single item report by ID
func GetItemByID(c *gin.Context) {
	id := c.Param("id")

	itemsMutex.RLock()
	defer itemsMutex.RUnlock()

	for _, item := range items {
		if item.ID == id {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    item,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Item tidak ditemukan dengan ID tersebut",
	})
}

// CreateItem creates a new lost or found item report
func CreateItem(c *gin.Context) {
	var newItem models.Item

	if err := c.ShouldBindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Validasi gagal: judul, status (LOST/FOUND), dan kontak wajib diisi",
			"details": err.Error(),
		})
		return
	}

	// Validate status enum
	if newItem.Status != "LOST" && newItem.Status != "FOUND" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Status harus bernilai 'LOST' atau 'FOUND'",
		})
		return
	}

	itemsMutex.Lock()
	newItem.ID = fmt.Sprintf("%d", nextID)
	nextID++
	newItem.CreatedAt = time.Now()
	items = append(items, newItem)
	itemsMutex.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Laporan barang berhasil dibuat",
		"data":    newItem,
	})
}
