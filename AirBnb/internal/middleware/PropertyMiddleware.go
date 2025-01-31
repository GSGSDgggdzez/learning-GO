package middleware

import (
	"AirBnb/internal/database"
	"AirBnb/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func PropertyOwner(s database.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract the property ID from the request
		propertyID, err := c.ParamsInt("id")
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid property ID", nil)
		}

		// Fetch the property from the database
		property, err := s.FindPropertyById(propertyID)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch property", err.Error())
		}

		// Check if the property exists
		if property == nil {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Property not found", nil)
		}

		// Extract the user ID from the JWT claims
		claims := c.Locals("user").(*utils.Claims)
		userID := claims.UserID

		// Verify that the user is the owner of the property
		if property.LandlordID != uint(userID) {
			return utils.SendErrorResponse(c, fiber.StatusForbidden, "You are not the owner of this property", nil)
		}

		// If the user is the owner, proceed to the next handler
		return c.Next()
	}
}
