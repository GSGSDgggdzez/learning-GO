package server

import (
	"chat/internal/controllers"
	middleware "chat/internal/middlewares"
	"chat/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func (s *FiberServer) RegisterFiberRoutes() {
	// Apply CORS middleware
	s.App.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: false, // credentials require explicit origins
		MaxAge:           300,
	}))

	authController := controllers.NewAuthController(s.db)
	GroupController := controllers.NewGroupController(s.db)
	auth := s.App.Group("/auth")
	Api := s.App.Group("/api")

	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Get("/verify", authController.VerifyEmail)
	auth.Post("/ResetPassword", authController.ResetPassword)

	// Update the route to handle JSON
	auth.Post("/ForgotPassword", authController.ForgotPassword)

	Api.Use(middleware.AuthRequired())
	Api.Get("/", s.InitialHandler)
	Api.Delete("/auth/delete", authController.DeleteUser)
	Api.Get("/groups", GroupController.GetAllGroups)
	Api.Post("/group/create", GroupController.Create)
	Api.Get("/group/find/:id", GroupController.FindGroup)
	Api.Delete("/group/delete/:id", GroupController.DeleteGroup)

	s.App.Get("/", s.HelloWorldHandler)
	s.App.Get("/health", s.healthHandler)

}

func (s *FiberServer) InitialHandler(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Email updated successfully",
		"user": fiber.Map{
			"id":         claims.UserID,
			"email":      claims.Email,
			"Name":       claims.Name,
			"profileUrl": claims.Avatar,
		},
	})
}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "Hello World",
	}

	return c.JSON(resp)
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}
