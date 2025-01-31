package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name         string        `gorm:"not null;size:255"`
	Email        string        `gorm:"not null;size:255;unique"`
	Password     string        `gorm:"not null;size:255"`
	Avatar       string        `gorm:"size:255"`
	Token        string        `gorm:"not null;size:255"`
	IsVerified   bool          `gorm:"default:false"`
	Resume       string        `gorm:"size:255"` // URL to resume
	Skills       []Skill       `gorm:"many2many:user_skills;"`
	Applications []Application `gorm:"foreignKey:UserID"`
	IsEmployer   bool          `gorm:"default:false"`
	CompanyID    *uint         // If user is an employer
	Company      *Company      `gorm:"foreignKey:CompanyID"`
}
