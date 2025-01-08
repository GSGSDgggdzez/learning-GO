package server

import (
	"github.com/gofiber/fiber/v2"

	"Video_to_text/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "Video_to_text",
			AppName:      "Video_to_text",
		}),

		db: database.New(),
	}

	return server
}
