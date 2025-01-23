package server

import (
	"github.com/gofiber/fiber/v2"

	"JobBoard/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "JobBoard",
			AppName:      "JobBoard",
		}),

		db: database.New(),
	}

	return server
}
