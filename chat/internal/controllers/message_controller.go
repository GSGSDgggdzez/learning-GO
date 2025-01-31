package controllers

import (
	"chat/internal/database"
	"chat/internal/models"
	"chat/internal/utils"
	"fmt"
	"mime/multipart"
	"time"

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
	Message    string `json:"message" validate:"required"`
	ReceiverId int    `json:"receiver_id" validate:"required"`
	GroupId    int    `json:"group_id"`
}

func (mc *MessageController) CreateMessage(c *fiber.Ctx) error {
	var req CreateMessageRequest
	claims := c.Locals("user").(*utils.Claims)

	// Parse form
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request body",
			"status": fiber.StatusBadRequest,
		})
	}

	// Create message
	messageData := database.MessageData{
		SenderId:   claims.UserID,
		ReceiverId: req.ReceiverId,
		Message:    req.Message,
		GroupId:    req.GroupId,
	}

	message, err := mc.db.CreateMessage(messageData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create message",
			"status": fiber.StatusInternalServerError,
		})
	}

	// Handle attachments if any
	form, err := c.MultipartForm()
	if err == nil && form.File["attachments"] != nil {
		for _, file := range form.File["attachments"] {
			attachment, err := mc.handleAttachment(file, message.Id)
			if err != nil {
				continue
			}
			message.Attachments = append(message.Attachments, attachment)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": message,
		"status":  fiber.StatusCreated,
	})
}

func (mc *MessageController) handleAttachment(file *multipart.FileHeader, messageId int) (*models.MessageAttachment, error) {
	// Save file
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	path := fmt.Sprintf("./uploads/%s", filename)

	if err := mc.SaveFile(file, path); err != nil {
		return nil, err
	}

	attachment := &models.MessageAttachment{
		MessageId: messageId,
		Name:      file.Filename,
		Path:      path,
		Mime:      file.Header.Get("Content-Type"),
		Size:      int(file.Size),
	}

	return mc.db.CreateMessageAttachment(attachment)
}
