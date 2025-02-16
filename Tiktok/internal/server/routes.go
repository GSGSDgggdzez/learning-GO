package server

import (
	controllers "Tiktok/internal/controller"
	"Tiktok/internal/middleware"

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

	// Initialize controllers
	postController := controllers.NewPostController(s.db) // 🎮 New post controller ready for action!

	authController := controllers.NewAuthController(s.db)

	auth := s.App.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Get("/verify/:token", authController.VerifyEmail)
	auth.Post("/forgot-password", authController.ForgotPassword)
	auth.Get("/reset-password/:Token", authController.RestPassword)

	// Protected API routes
	api := s.App.Group("/api", middleware.AuthRequired())

	// User management
	api.Delete("/auth/delete/:ID", authController.DeleteUser)

	// 📝 Post routes - let's make some noise!
	posts := api.Group("/posts")
	posts.Post("/create", postController.CreatePost) // 🎬 Create amazing new posts
	posts.Delete("/delete/:id", postController.DeletePost)
	posts.Put("/edit/:id", postController.UpdatePost)
	s.App.Get("/", s.HelloWorldHandler)

	s.App.Get("/health", s.healthHandler)

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
