package controllers

import (
	"Tiktok/internal/database"
	"Tiktok/internal/utils"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type FollowController struct {
	db       database.Service
	validate *validator.Validate
}

func NewFollowController(db database.Service) *FollowController {
	return &FollowController{
		db:       db,
		validate: validator.New(),
	}
}

func (fc *FollowController) FollowUser(c *fiber.Ctx) error {
	userToFollowID, err := c.ParamsInt("id")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
	}

	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or missing authentication", "")
	}

	// Check if already following
	existingFollow, err := fc.db.FindFollowByUsers(uint(claims.UserID), uint(userToFollowID))
	if err == nil && existingFollow != nil {
		// Unfollow
		if err := fc.db.DeleteFollow(existingFollow.ID); err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to unfollow user", err.Error())
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "User unfollowed successfully",
		})
	}

	// Create new follow relationship
	follow := database.Follow{
		FollowerID:  uint(claims.UserID),
		FollowingID: uint(userToFollowID),
	}

	result, err := fc.db.CreateFollow(follow)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to follow user", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User followed successfully",
		"follow":  result,
	})
}
