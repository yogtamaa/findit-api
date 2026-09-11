package models

import "time"

// Report represents a row in the `reports` table.
// A single "reports" table now holds both lost and found reports,
// distinguished by the Type field ("lost" or "found").
type Report struct {
	ID               uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	ReportIdentifier string     `gorm:"column:report_identifier;type:varchar(50);uniqueIndex" json:"report_identifier"`
	UserID           uint       `gorm:"column:user_id;not null" json:"user_id" binding:"required"`
	Type             string     `gorm:"type:varchar(20);not null" json:"type" binding:"required,oneof=lost found"`
	Title            string     `gorm:"type:varchar(255);not null" json:"title" binding:"required"`
	Description      string     `gorm:"type:text" json:"description"`
	Category         string     `gorm:"type:varchar(100)" json:"category"`
	Location         string     `gorm:"type:varchar(255)" json:"location"`
	PhotoURL         string     `gorm:"column:photo_url;type:varchar(255)" json:"photo_url"`
	Status           string     `gorm:"type:varchar(30);default:'baru'" json:"status"`
	ActivityNote     string     `gorm:"column:activity_note;type:varchar(255)" json:"activity_note"`
	ItemDate         *time.Time `gorm:"column:item_date" json:"item_date"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (Report) TableName() string {
	return "reports"
}
