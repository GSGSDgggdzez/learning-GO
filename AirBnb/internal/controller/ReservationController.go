package controllers

import (
	"AirBnb/internal/database"

	"github.com/go-playground/validator"
)

type ReservationController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

func NewReservationController(db database.Service) *ReservationController {
	return &ReservationController{
		db:       db,
		validate: validator.New(),
	}
}
