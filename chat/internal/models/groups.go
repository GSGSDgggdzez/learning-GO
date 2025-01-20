package models

import "time"

type Group struct {
	Id            int       `gorm:"primaryKey;autoIncrement"`
	Name          string    `gorm:"not null;size:255"`
	Description   string    `gorm:"size:255"`
	Image         string    `gorm:"size:255"`
	OwnerId       int       `gorm:"not null;foreignKey:users(id)"`
	LastMassageId int       `gorm:"foreignKey:messages(id)"`
	CreateAt      time.Time `gorm:"autoCreateTime"`
}
