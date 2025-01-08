package models

import (
	"time"
)

// User represents a user in the system.
type User struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"` // Primary key with auto increment
	Name           string    `gorm:"size:100;not null"`        // First name of the user
	Email          string    `gorm:"size:100;unique;not null"` // Email address (unique and not null)
	Password       string    `gorm:"size:255;not null"`        // Password (hashed)
	Profile_Url    string    `gorm:"size:255;null"`
	Email_Verified string    `gorm:"size:255;null"`
	Token          string    `gorm:"size:255;null"`
	CreatedAt      time.Time // Timestamp of when the user was created
	UpdatedAt      time.Time // Timestamp of the last update
}
