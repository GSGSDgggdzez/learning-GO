package models

import (
	"time"
)

type User struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	Email      string `gorm:"uniqueIndex;not null"`
	Password   string `gorm:"not null" json:"-"`
	Name       string `gorm:"not null;size:255"`
	Avatar     string `gorm:"not null;size:255"`
	Token      string `gorm:"not null;size:255"`
	IsVerified bool   `gorm:"default:false"`
	IsActive   bool   `gorm:"default:true"`
	IsStaff    bool   `gorm:"default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Properties   []Property    `gorm:"foreignKey:LandlordID"`
	Reservations []Reservation `gorm:"foreignKey:CreatedByID"`
	Favorites    []Property    `gorm:"many2many:user_favorites;"`
}
