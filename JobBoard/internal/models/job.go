package models

import (
	"time"

	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Title           string  `gorm:"not null;size:255"`
	Description     string  `gorm:"type:text;not null"`
	CompanyID       uint    `gorm:"not null"`
	Company         Company `gorm:"foreignKey:CompanyID"`
	Location        string  `gorm:"size:255"`
	SalaryMin       float64 `gorm:"type:decimal(10,2)"`
	SalaryMax       float64 `gorm:"type:decimal(10,2)"`
	Currency        string  `gorm:"size:3;default:'EUR'"` // EUR, USD, etc.
	JobType         string  `gorm:"size:50"`              // Full-time, Part-time, Contract, etc.
	ExperienceLevel string  `gorm:"size:50"`              // Junior, Mid, Senior, etc.
	Remote          bool    `gorm:"default:false"`
	Skills          []Skill `gorm:"many2many:job_skills;"`
	Applications    []Application
	ExpiresAt       time.Time
}
