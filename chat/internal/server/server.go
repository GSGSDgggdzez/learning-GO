package server

import (
	"github.com/gofiber/fiber/v2"

	"chat/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "chat",
			AppName:      "chat",
		}),

		db: database.New(),
	}

	return server
}
