package models

import "time"

type Comment struct {
	ID        uint `gorm:"primaryKey;;autoIncrement"`
	UserID    uint
	PostID    uint
	User      User `gorm:"foreignKey:UserID"`
	Post      Post `gorm:"foreignKey:PostID"`
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
