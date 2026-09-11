package models

import "time"

// Item represents a lost or found item report
type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Status      string    `json:"status" binding:"required"` // "LOST" or "FOUND"
	Location    string    `json:"location"`
	Contact     string    `json:"contact" binding:"required"`
	CreatedAt   time.Time `json:"created_at"`
}
