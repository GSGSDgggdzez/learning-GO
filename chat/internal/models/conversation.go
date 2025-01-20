package models

import "time"

type Conversation struct {
	Id            int       `gorm:"primaryKey;autoIncrement"`
	UserId1       int       `gorm:"not null;foreignKey:users(id)"`
	UserId2       int       `gorm:"not null;foreignKey:users(id)"`
	LastMassageId int       `gorm:"foreignKey:messages(id)"`
	CreateAt      time.Time `gorm:"autoCreateTime"`
}
