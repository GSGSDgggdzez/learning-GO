package server

import (
	"github.com/gofiber/fiber/v2"

	"Learning_graphgl/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "Learning_graphgl",
			AppName:      "Learning_graphgl",
		}),

		db: database.New(),
	}

	return server
}
