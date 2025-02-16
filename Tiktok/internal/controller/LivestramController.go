package controllers

import (
	"Tiktok/internal/database"
	"Tiktok/internal/models"
	"Tiktok/internal/utils"
	"context"
	"strconv"

	"github.com/GetStream/getstream-go"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LiveStreamController struct {
	db           database.Service
	validate     *validator.Validate
	streamClient *getstream.Client
}

type StartStreamRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (lc *LiveStreamController) StartStream(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("user").(*utils.Claims)

	var req StartStreamRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request", err.Error())
	}

	userID := strconv.Itoa(claims.UserID)
	callID := uuid.New().String()

	// Create livestream call
	call := lc.streamClient.Video().Call("livestream", callID)

	response, err := call.GetOrCreate(ctx, &getstream.GetOrCreateCallRequest{
		Data: &getstream.CallRequest{
			CreatedByID: getstream.PtrTo(userID),
			Members: []getstream.MemberRequest{
				{UserID: userID, Role: getstream.PtrTo("host")},
			},
			Custom: map[string]interface{}{
				"title":       req.Title,
				"description": req.Description,
			},
		},
	})

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create stream", err.Error())
	}

	// Save to database
	stream := &models.LiveStream{
		UserID:      claims.UserID,
		Title:       req.Title,
		Description: req.Description,
		CallID:      callID,
		Status:      "backstage",
		StreamKey:   response.StreamKey,
		RtmpUrl:     response.RtmpUrl,
	}

	if err := lc.db.CreateLiveStream(ctx, stream); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to save stream", err.Error())
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   stream,
	})
}
