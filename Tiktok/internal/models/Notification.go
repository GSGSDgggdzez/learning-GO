package models

import "time"

type Notification struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	UserID    uint   // recipient
	FromID    uint   // sender
	User      User   `gorm:"foreignKey:UserID"`
	From      User   `gorm:"foreignKey:FromID"`
	Type      string `gorm:"size:50"` // like, comment, follow, gift
	Content   string `gorm:"size:255"`
	Read      bool   `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
