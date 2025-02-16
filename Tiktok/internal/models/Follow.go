package models

import "time"

type Follow struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	FollowerID  uint // User who follows
	FollowingID uint // User being followed
	Follower    User `gorm:"foreignKey:FollowerID"`
	Following   User `gorm:"foreignKey:FollowingID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
