package models

import "time"

type Post struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	UserID    uint
	User      User   `gorm:"foreignKey:UserID"`
	Text      string `gorm:"size:255"`
	Video     string `gorm:"size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Comments  []Comment `gorm:"foreignKey:PostID"`
	Likes     []Like    `gorm:"foreignKey:PostID"`
	// Add these fields
	Hashtags   []Hashtag `gorm:"many2many:post_hashtags;"`
	ViewCount  uint      `gorm:"default:0"`
	ShareCount uint      `gorm:"default:0"`
	SaveCount  uint      `gorm:"default:0"`
	Duration   float64   // Video duration in seconds
	IsPrivate  bool      `gorm:"default:false"`
	Music      string    `gorm:"size:255"` // Background music/sound
	Location   string    `gorm:"size:255"`
}
