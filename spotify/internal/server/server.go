package server

import (
	"github.com/gofiber/fiber/v2"

	"spotify/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "spotify",
			AppName:      "spotify",
		}),

		db: database.New(),
	}

	return server
}
