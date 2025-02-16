package models

import "time"

type ConversationMessage struct {
	ID             uint   `gorm:"primaryKey;autoIncrement"`
	Body           string `gorm:"type:text"`
	CreatedAt      time.Time
	ConversationID uint
	CreatedByID    uint
	SentToID       uint
	Conversation   Conversation `gorm:"foreignKey:ConversationID"`
	CreatedBy      User         `gorm:"foreignKey:CreatedByID"`
	SentTo         User         `gorm:"foreignKey:SentToID"`
}
