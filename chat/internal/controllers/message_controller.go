package controllers

import (
	"chat/internal/database"
	"chat/internal/utils"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type MessageController struct {
	db       database.Service
	validate *validator.Validate
}

func NewMessageController(db database.Service) *MessageController {
	return &MessageController{
		db:       db,
		validate: validator.New(),
	}
}

type CreateMessageRequest struct {
	Message string `json:"message" validate:"required" `
}

func (mc *MessageController) CreateMessage(c *fiber.Ctx) error {
	var req CreateMessageRequest

	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
	}

	if err := mc.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": utils.FormatValidationErrors(err),
			"status":  fiber.StatusBadRequest,
		})
	}
}
