package database

import (
	"chat/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.Conversation{},
		&models.Message{},
		&models.MessageAttachment{},
	)
}
