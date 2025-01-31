package models

import "time"

type Property struct {
	ID            uint   `gorm:"primaryKey;autoIncrement"`
	Title         string `gorm:"not null"`
	Description   string `gorm:"type:text"`
	PricePerNight int    `gorm:"not null"`
	Bedrooms      int
	Bathrooms     int
	Guests        int
	Country       string
	CountryCode   string
	Category      string
	Image         string
	CreatedAt     time.Time
	LandlordID    uint
	Landlord      User          `gorm:"foreignKey:LandlordID"`
	FavoritedBy   []User        `gorm:"many2many:user_favorites;"`
	Reservations  []Reservation `gorm:"foreignKey:PropertyID"`
}
