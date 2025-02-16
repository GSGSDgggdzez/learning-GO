package server

import (
	"github.com/gofiber/fiber/v2"

	"Tiktok/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "Tiktok",
			AppName:      "Tiktok",
		}),

		db: database.New(),
	}

	return server
}
