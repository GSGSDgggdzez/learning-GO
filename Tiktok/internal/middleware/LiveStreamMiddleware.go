package middleware

import (
	"Tiktok/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func CreateLiveStream() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("user").(*utils.Claims)
		if !ok || claims == nil {
			return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or missing authentication", "")
		}

		// Check if user has more than 1000 followers
		if claims.FollowerCount < 1000 {
			return utils.SendErrorResponse(c, fiber.StatusForbidden, "You need at least 1000 followers to start a live stream", "")
		}

		return c.SendStatus(fiber.StatusOK)
	}
}
