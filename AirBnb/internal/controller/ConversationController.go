package controllers

import (
	"AirBnb/internal/database"
	"AirBnb/internal/utils"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type ConversationController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

func NewConversationController(db database.Service) *ConversationController {
	return &ConversationController{
		db:       db,
		validate: validator.New(),
	}
}

// ----------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the registration logic -------------------------
// ----------------------------------------------------------------------------------------------------

type ConversationRequest struct {
	Message string `json:"message" validate:"required"`
}

func (cc *ConversationController) CreateConversationMessage(c *fiber.Ctx) error {
	// Extract user claims from the context
	claims := c.Locals("user").(*utils.Claims)

	// Extract the receiver's ID from the URL parameters
	receiverID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid receiver ID", err.Error())
	}

	// Parse the request body
	var req ConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	// Validate the request struct
	if err := cc.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed, please check input", utils.FormatValidationErrors(err))
	}

	// Find or create a conversation between the sender and receiver
	conversation, err := cc.db.FindOrCreateConversation(uint(claims.UserID), uint(receiverID))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to find or create conversation", err.Error())
	}

	// Create the message
	message := database.ConversationMessage{
		Body:           req.Message,
		ConversationID: conversation.ID,
		CreatedByID:    uint(claims.UserID),
		// SentToID:       uint(receiverID),
	}

	// Save the message to the database
	if err := cc.db.CreateConversationMessage(message); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create message", err.Error())
	}

	// Return a success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Message created successfully",
		"data":    message,
	})
}
