package server

import (
	"Video_to_text/internal/controllers" // Ensure this matches your project structure
	"Video_to_text/internal/middleware"

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

	// Initialize auth controller
	authController := controllers.NewAuthController(s.db)
	PostController := controllers.NewPostController(s.db)

	// Auth routes
	auth := s.App.Group("/auth")
	Api := s.App.Group("/api")
	// Public routes
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Get("/verify", authController.VerifyEmail)
	auth.Get("/ResetPassword", authController.ResetPassword)

	// Protected routes
	auth.Use(middleware.AuthRequired())
	auth.Get("/ForgotPassword", authController.ResetPassword)
	auth.Post("/updateUser", authController.UpdateUser)
	auth.Post("/deleteUser", authController.DeleteUser)
	Api.Get("/Url", PostController.CreatePost)

	// Existing routes...
	s.App.Get("/", s.HelloWorldHandler)
	s.App.Get("/health", s.healthHandler)

}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "Hello World",
	}

	return c.JSON(resp)
}

func (s *FiberServer) RegisterHandeler(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "Hello World",
	}

	return c.JSON(resp)
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}
