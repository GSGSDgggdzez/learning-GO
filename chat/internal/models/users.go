package models

import "time"

type User struct {
	Id             int       `gorm:"primaryKey;autoIncrement"`
	Name           string    `gorm:"not null;size:255"`
	Email          string    `gorm:"not null;size:255;unique"`
	Password       string    `gorm:"not null;size:255"`
	Avatar         string    `gorm:"size:255"`
	Token          string    `gorm:"size:255"`
	Email_verified string    `gorm:"null"`
	Is_admin       bool      `gorm:"default:false"`
	CreateAt       time.Time `gorm:"autoCreateTime"`
}
