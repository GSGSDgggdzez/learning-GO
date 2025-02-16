package middleware

import (
	"AirBnb/internal/database"
	"AirBnb/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func ReservationOwner(s database.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract the reservation ID from the request
		reservationID, err := c.ParamsInt("id")
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid reservation ID", nil)
		}

		// Fetch the reservation from database
		reservation, err := s.FindReservationById(uint(reservationID))
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reservation", err.Error())
		}

		// Check if the reservation exists
		if reservation == nil {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Reservation not found", nil)
		}

		// Extract the user ID from the JWT claims
		claims := c.Locals("user").(*utils.Claims)
		userID := claims.UserID

		// Verify that the user owns this reservation
		if reservation.CreatedByID != uint(userID) {
			return utils.SendErrorResponse(c, fiber.StatusForbidden, "You are not authorized to access this reservation", nil)
		}

		// If the user is the owner, proceed to the next handler
		return c.Next()
	}
}
