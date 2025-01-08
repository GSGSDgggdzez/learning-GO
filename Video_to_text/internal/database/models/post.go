package models

import (
	"time"
)

type Post struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"not null"`          // Changed type from User to uint
	User      User      `gorm:"foreignKey:UserID"` // Added proper relationship reference
	Url       string    `gorm:"size:255;not null"`
	Title     string    `gorm:"size:255;not null"`
	Body      string    `gorm:"size:255;not null"`
	CreatedAt time.Time // Timestamp of when the user was created
	UpdatedAt time.Time // Timestamp of the last update
}
