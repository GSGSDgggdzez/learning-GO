package middleware

import (
	"Tiktok/internal/utils"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		// Add debug logging
		fmt.Printf("Auth Header: %s\n", authHeader)

		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":  "Missing Authorization header",
				"status": fiber.StatusUnauthorized,
			})
		}

		// Verify Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":  "Invalid Authorization format. Use 'Bearer <token>'",
				"status": fiber.StatusUnauthorized,
			})
		}

		claims, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":  "Invalid token",
				"status": fiber.StatusUnauthorized,
			})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}
