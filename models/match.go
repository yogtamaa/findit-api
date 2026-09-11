package models

import "time"

// Match represents a row in the `matches` table, linking a lost report
// with a found report once the AI/manual matching process pairs them.
type Match struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	LostReportID    uint       `gorm:"column:lost_report_id;not null" json:"lost_report_id" binding:"required"`
	FoundReportID   uint       `gorm:"column:found_report_id;not null" json:"found_report_id" binding:"required"`
	Status          string     `gorm:"type:varchar(30);default:'pending'" json:"status"`
	VerifiedBy      *uint      `gorm:"column:verified_by" json:"verified_by"`
	HandoverMethod  string     `gorm:"column:handover_method;type:varchar(50)" json:"handover_method"`
	ContactSharedAt *time.Time `gorm:"column:contact_shared_at" json:"contact_shared_at"`
	ActivityNote    string     `gorm:"column:activity_note;type:varchar(255)" json:"activity_note"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (Match) TableName() string {
	return "matches"
}
