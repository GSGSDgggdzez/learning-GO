package controllers

import (
	"AirBnb/internal/database"
	"AirBnb/internal/utils"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type FavoritesController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

func NewFavoritesController(db database.Service) *FavoritesController {
	return &FavoritesController{
		db:       db,
		validate: validator.New(),
	}
}

// ----------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the registration logic -------------------------
// ----------------------------------------------------------------------------------------------------

func (fc *FavoritesController) AddToFavorites(c *fiber.Ctx) error {
	// Extract user claims from the context
	claims := c.Locals("user").(*utils.Claims)

	propertyID := c.Params("id")

	id, err := strconv.Atoi(propertyID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}

	// Call the database service to add the property to favorites
	property, err := fc.db.AddFavoriteProperty(uint(id), uint(claims.UserID))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": fiber.StatusOK,
		"data":   property,
	})
}

// ----------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the registration logic -------------------------
// ----------------------------------------------------------------------------------------------------

func (fc *FavoritesController) DeleteFromFavorites(c *fiber.Ctx) error {
	// Extract user claims from the context
	claims := c.Locals("user").(*utils.Claims)

	propertyID := c.Params("id")

	id, err := strconv.Atoi(propertyID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}

	property, err := fc.db.DeleteFavoriteProperty(uint(id), uint(claims.UserID))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": fiber.StatusOK,
		"data":   property,
	})
}
