package controllers

import (
	"AirBnb/internal/database"
	"AirBnb/internal/utils"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
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

type CreateReservationRequest struct {
	StartDate      time.Time `json:"start_date" validate:"required,future"`
	EndDate        time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	NumberOfNights int       `json:"number_of_nights" validate:"required,min=1"`
	Guests         int       `json:"guests" validate:"required,min=1"`
}

func (rc *ReservationController) CreateReservation(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	propertyID := c.Params("id")

	// Convert the ID to an integer
	id, err := strconv.Atoi(propertyID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}

	var req CreateReservationRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := rc.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	property, err := rc.db.FindPropertyById(int(id))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Property not found", nil)
	}

	totalPrice := float64(property.PricePerNight * req.NumberOfNights)

	createReservationData := database.Reservation{
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		NumberOfNights: req.NumberOfNights,
		Guests:         req.Guests,
		TotalPrice:     totalPrice,
		CreatedByID:    uint(claims.UserID),
		PropertyID:     uint(id),
	}

	createdReservation, err := rc.db.CreateReservation(createReservationData)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create reservation", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Reservation successfully created",
		"property": fiber.Map{
			"property": createdReservation,
		},
	})
}

func (rc *ReservationController) DeleteReservation(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	reservationID := c.Params("id")
	id, err := strconv.Atoi(reservationID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid reservation ID", nil)
	}

	// Get the reservation first to check dates
	reservation, err := rc.db.FindReservationById(uint(id))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Reservation not found", nil)
	}

	// Check if user owns this reservation
	if reservation.CreatedByID != uint(claims.UserID) {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Not authorized to delete this reservation", nil)
	}

	// Check if reservation starts within 24 hours
	if time.Until(reservation.StartDate) < 24*time.Hour {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Cannot cancel reservations less than 24 hours before start date", nil)
	}

	// Proceed with deletion
	deletedReservation, err := rc.db.DeleteReservation(uint(id))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete reservation", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Reservation successfully deleted",
		"property": fiber.Map{
			"property": deletedReservation,
		},
	})
}
