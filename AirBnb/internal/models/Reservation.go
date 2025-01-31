package models

import "time"

type Reservation struct {
	ID             uint `gorm:"primaryKey;autoIncrement"`
	StartDate      time.Time
	EndDate        time.Time
	NumberOfNights int
	Guests         int
	TotalPrice     float64
	CreatedAt      time.Time
	CreatedByID    uint
	PropertyID     uint
	CreatedBy      User     `gorm:"foreignKey:CreatedByID"`
	Property       Property `gorm:"foreignKey:PropertyID"`
}
