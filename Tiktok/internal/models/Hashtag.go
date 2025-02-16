package models

import "time"

type Hashtag struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"unique;size:255"`
	Posts     []Post `gorm:"many2many:post_hashtags;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
