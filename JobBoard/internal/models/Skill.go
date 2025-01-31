package models

type Skill struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null;size:100;unique"`
	Jobs []Job  `gorm:"many2many:job_skills;"`
}
