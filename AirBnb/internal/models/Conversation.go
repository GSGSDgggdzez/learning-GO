package models

import "time"

type Conversation struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	CreatedAt  time.Time
	ModifiedAt time.Time
	Users      []User `gorm:"many2many:conversation_users;"`
	Messages   []ConversationMessage
}
