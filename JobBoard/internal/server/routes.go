package server

import (
	controllers "JobBoard/internal/controller"
	middleware "JobBoard/internal/middleware"

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
	companyController := controllers.NewCompanyController(s.db)
	auth := s.App.Group("/auth")

	auth.Post("/register", authController.Register)
	auth.Get("/verify", authController.VerifyEmail)
	s.App.Get("/", s.HelloWorldHandler)

	s.App.Get("/health", s.healthHandler)

	// Company routes with authentication middleware
	company := s.App.Group("/company", middleware.AuthRequired())
	company.Post("/register", companyController.RegisterCompany)
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
