package models

import "time"

type LiveStreamComment struct {
	ID           uint `gorm:"primaryKey;autoIncrement"`
	UserID       uint
	LiveStreamID uint
	User         User       `gorm:"foreignKey:UserID"`
	LiveStream   LiveStream `gorm:"foreignKey:LiveStreamID"`
	Text         string     `gorm:"size:255"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
