package server

import (
	"github.com/gofiber/fiber/v2"

	"AirBnb/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "AirBnb",
			AppName:      "AirBnb",
		}),

		db: database.New(),
	}

	return server
}
