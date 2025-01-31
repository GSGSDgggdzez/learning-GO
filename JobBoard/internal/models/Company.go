package models

import "gorm.io/gorm"

type Company struct {
	gorm.Model
	Name        string `gorm:"not null;size:255"`
	Email       string `gorm:"not null;size:255;unique"`
	Logo        string `gorm:"size:255"` // URL to logo image
	Description string `gorm:"type:text"`
	Website     string `gorm:"size:255"`
	Location    string `gorm:"size:255"`
	Jobs        []Job  `gorm:"foreignKey:CompanyID"`
}
