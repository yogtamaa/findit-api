package models

import "time"

// User represents a row in the `users` table.
type User struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string     `gorm:"type:varchar(150);not null" json:"name" binding:"required"`
	Email           string     `gorm:"type:varchar(150);uniqueIndex;not null" json:"email" binding:"required,email"`
	Phone           string     `gorm:"type:varchar(30)" json:"phone"`
	Role            string     `gorm:"type:varchar(30);default:'user'" json:"role"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	// Password is only used for binding on register/login requests, never sent back in a response.
	Password      string    `gorm:"type:varchar(255);not null" json:"-"`
	RememberToken string    `gorm:"type:varchar(100)" json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
