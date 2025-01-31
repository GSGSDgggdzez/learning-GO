package models

import "gorm.io/gorm"

type Application struct {
	gorm.Model
	JobID       uint   `gorm:"not null"`
	Job         Job    `gorm:"foreignKey:JobID"`
	UserID      uint   `gorm:"not null"`
	User        User   `gorm:"foreignKey:UserID"`
	CoverLetter string `gorm:"type:text"`
	ResumeURL   string `gorm:"size:255"`
	Status      string `gorm:"size:50;default:'pending'"` // pending, accepted, rejected
}
