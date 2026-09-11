package models

import "time"

// FoundItem struct representasi dari tabel found_items
type FoundItem struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint       `gorm:"column:user_id;default:0" json:"user_id"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title" binding:"required"`
	Description string     `gorm:"type:text" json:"description"`
	Category    string     `gorm:"type:varchar(100)" json:"category"`
	Location    string     `gorm:"type:varchar(255)" json:"location"`
	Contact     string     `gorm:"type:varchar(100)" json:"contact"`
	Status      string     `gorm:"type:varchar(50);default:'FOUND'" json:"status"`
	Date        *time.Time `gorm:"type:date" json:"date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName menentukan nama tabel di MySQL
func (FoundItem) TableName() string {
	return "found_items"
}
