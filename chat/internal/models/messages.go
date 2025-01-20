package models

import "time"

type Message struct {
	Id         int       `gorm:"primaryKey;autoIncrement"`
	SenderId   int       `gorm:"not null;foreignKey:users(id)"`
	Message    string    `gorm:"not null;"`
	ReceiverId int       `gorm:"foreignKey:users(id)"`
	GroupId    int       `gorm:"foreignKey:groups(id)"`
	CreateAt   time.Time `gorm:"autoCreateTime"`
}
