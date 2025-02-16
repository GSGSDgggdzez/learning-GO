package models

import "time"

type Gift struct {
	ID           uint `gorm:"primaryKey;autoIncrement"`
	UserID       uint // sender
	LiveStreamID uint
	User         User       `gorm:"foreignKey:UserID"`
	LiveStream   LiveStream `gorm:"foreignKey:LiveStreamID"`
	GiftType     string     `gorm:"size:50"`
	Amount       float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
