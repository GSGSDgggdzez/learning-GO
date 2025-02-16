package models

import "time"

type Like struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	UserID    uint
	PostID    uint
	User      User `gorm:"foreignKey:UserID"`
	Post      Post `gorm:"foreignKey:PostID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
