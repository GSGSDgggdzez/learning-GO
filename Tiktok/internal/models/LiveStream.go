package models

import "time"

type LiveStream struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	UserID      uint
	User        User   `gorm:"foreignKey:UserID"`
	Title       string `gorm:"size:255"`
	Description string `gorm:"size:1000"`
	StreamKey   string `gorm:"unique;size:255"`
	Status      string `gorm:"default:'offline'"` // offline, live, ended
	ViewerCount uint   `gorm:"default:0"`
	IsPublic    bool   `gorm:"default:false"`
	RtmpUrl     string `gorm:"size:255"`
	CallID      string // Stream's call ID
	StartedAt   time.Time
	EndedAt     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Comments    []LiveStreamComment `gorm:"foreignKey:LiveStreamID"`
	Gifts       []Gift              `gorm:"foreignKey:LiveStreamID"`
}
