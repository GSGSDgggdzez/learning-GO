package models

import "time"

type User struct {
	ID             uint   `gorm:"primaryKey;autoIncrement"`
	Name           string `gorm:"not null;size:255"`
	Bio            string `gorm:"null;size:255"`
	Avatar         string `gorm:"size:255"`
	Email          string `gorm:"unique"`
	EmailVerified  bool   `gorm:"default:false"`
	Password       string `gorm:"not null" json:"-"`
	Token          string `gorm:"not null;size:255"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Posts          []Post         `gorm:"foreignKey:UserID"`
	Comments       []Comment      `gorm:"foreignKey:UserID"`
	Likes          []Like         `gorm:"foreignKey:UserID"`
	FollowerCount  uint           `gorm:"default:0"`
	FollowingCount uint           `gorm:"default:0"`
	Coins          float64        `gorm:"default:0"` // Virtual currency for gifts
	IsVerified     bool           `gorm:"default:false"`
	LiveStreams    []LiveStream   `gorm:"foreignKey:UserID"`
	Followers      []Follow       `gorm:"foreignKey:FollowingID"`
	Following      []Follow       `gorm:"foreignKey:FollowerID"`
	Notifications  []Notification `gorm:"foreignKey:UserID"`
}
