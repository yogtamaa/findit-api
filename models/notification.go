package models

import "time"

// Notification represents a row in the `notifications` table.
// Note: unlike the other tables, `id` here is a UUID string (not an
// auto-increment int) — this matches the schema in the ERD where
// notifications.id is typed as text, not int.
type Notification struct {
	ID             string     `gorm:"primaryKey;type:char(36)" json:"id"`
	Type           string     `gorm:"type:varchar(150);not null" json:"type" binding:"required"`
	NotifiableType string     `gorm:"column:notifiable_type;type:varchar(150);not null" json:"notifiable_type" binding:"required"`
	NotifiableID   uint       `gorm:"column:notifiable_id;not null" json:"notifiable_id" binding:"required"`
	Data           string     `gorm:"type:text" json:"data"`
	ReadAt         *time.Time `gorm:"column:read_at" json:"read_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
