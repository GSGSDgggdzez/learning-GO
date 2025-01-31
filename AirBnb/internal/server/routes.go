package server

import (
	controllers "AirBnb/internal/controller"
	"AirBnb/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func (s *FiberServer) RegisterFiberRoutes() {
	s.App.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Initialize controllers
	authController := controllers.NewAuthController(s.db)
	propertiesController := controllers.NewPropertiesController(s.db)
	favoritesController := controllers.NewFavoritesController(s.db)

	// Auth routes (public)
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

	// Property routes with authentication
	properties := api.Group("/properties")
	properties.Post("/register", propertiesController.RegisterProperty)
	// In your routes configuration
	properties.Get("/", propertiesController.GetAllProperties)
	properties.Get("/:id", propertiesController.GetPropertyById)

	// Property routes with both auth and owner verification
	propertyProtected := properties.Group("/:id", middleware.PropertyOwner(s.db))
	propertyProtected.Delete("/", propertiesController.DeleteProperty)
	propertyProtected.Put("/", propertiesController.UpdateProperty)

	favorites := api.Group("/favorites")
	favorites.Post("/:id", favoritesController.AddToFavorites)
	favorites.Delete("/:id", favoritesController.DeleteFromFavorites)

	// Health check and root routes
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
