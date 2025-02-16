package controllers

import (
	"Tiktok/internal/database"
	"Tiktok/internal/utils"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type LikeController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

func NewLikController(db database.Service) *LikeController {
	return &LikeController{
		db:       db,              // Setting the provided database service.
		validate: validator.New(), // Initializing a new validator instance.
	}
}

type LikePostRequest struct {
	Like uint `json:"like" validate:"required,min=1,max=1"`
}

func (lc *LikeController) LikeVideos(c *fiber.Ctx) error {
	postID, err := c.ParamsInt("id")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid post ID", err.Error())
	}

	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or missing authentication", "")
	}

	var req LikePostRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request data", err.Error())
	}

	if err := lc.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	like := database.Like{
		UserID: uint(claims.UserID),
		PostID: uint(postID),
	}

	// Check if like already exists
	existingLike, err := lc.db.FindLikeByUserAndPost(like.UserID, like.PostID)
	if err == nil && existingLike != nil {
		// Unlike the video by removing the like
		if err := lc.db.DeleteLike(existingLike.ID); err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to unlike video", err.Error())
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Video unliked successfully",
		})
	}

	// Create new like
	createdLike, err := lc.db.CreateLike(like)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to like video", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Video liked successfully",
		"like":    createdLike,
	})
}
