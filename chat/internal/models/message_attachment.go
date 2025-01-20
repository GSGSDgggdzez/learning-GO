package models

import "time"

type MessageAttachment struct {
	Id        int       `gorm:"primaryKey;autoIncrement"`
	MessageId int       `gorm:"not null;foreignKey:messages(id)"`
	Name      string    `gorm:"not null;size:255"`
	Path      string    `gorm:"not null;size:255"`
	Mime      string    `gorm:"not null;size:255"`
	Size      int       `gorm:"not null"`
	CreateAt  time.Time `gorm:"autoCreateTime"`
}
