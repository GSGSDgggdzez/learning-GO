package controllers

import (
	"Video_to_text/internal/database"
	"Video_to_text/internal/utils"
	"html"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type PostController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

func NewPostController(db database.Service) *PostController {
	return &PostController{
		db:       db,              // Setting the provided database service.
		validate: validator.New(), // Initializing a new validator instance.
	}
}

type URLRequest struct {
	URL   string `json:"url" validate:"required,url"`
	Title string `json:"title" validate:"required,max=255"`
	Body  string `json:"body" validate:"required,max=255"`
}

func (Pc *PostController) CreatePost(c *fiber.Ctx) error {
	var req URLRequest

	claims := c.Locals("user").(*utils.Claims)
	// Parse the request body into the URLRequest struct
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
	}

	if err := Pc.validate.Struct(req); err != nil {
		// Return an error response if validation fails.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",               // Error message.
			"details": utils.FormatValidationErrors(err), // Formatting and sending detailed validation errors.
			"status":  fiber.StatusBadRequest,            // Unauthorized status code.
		})
	}

	CreatePost := database.PostCreate{
		Url:    html.EscapeString(req.URL),
		Title:  html.EscapeString(req.Title),
		Body:   html.EscapeString(req.Body),
		UserId: claims.UserID,
	}

	NewPost, err := Pc.db.CreatePost(CreatePost)

	if err != nil {
		// Return an error response if user creation fails.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create user",         // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Post created successfully", // Success message.
		"status":  fiber.StatusCreated,         // Created status code.
		"user": fiber.Map{
			"id":    NewPost.ID,
			"User":  NewPost.UserID,
			"Title": NewPost.Title,
			"Body":  NewPost.Body,
		},
	})

}

func (Pc *PostController) AllPost(c *fiber.Ctx) error {
	posts, err := Pc.db.AllPost(database.PostCreate{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve posts",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Posts retrieved successfully",
		"status":  fiber.StatusOK,
		"posts":   posts,
	})
}

type ParticularPostRequest struct {
	ID uint `json:"ID" validate:"required, max=255"`
}

func (Pc *PostController) ParticularPost(c *fiber.Ctx) error {
	var req ParticularPostRequest

	// Parse the request body into the URLRequest struct
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
	}

	if err := Pc.validate.Struct(req); err != nil {
		// Return an error response if validation fails.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",               // Error message.
			"details": utils.FormatValidationErrors(err), // Formatting and sending detailed validation errors.
			"status":  fiber.StatusBadRequest,            // Unauthorized status code.
		})
	}

	ParticularPost, err := Pc.db.FindPostById(database.PostCreate{Id: req.ID})

	if err != nil {
		// Return an error response if user creation fails.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create user",         // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Post created successfully", // Success message.
		"status":  fiber.StatusCreated,         // Created status code.
		"user": fiber.Map{
			"id":    ParticularPost.ID,
			"User":  ParticularPost.UserID,
			"Title": ParticularPost.Title,
			"Body":  ParticularPost.Body,
		},
	})

}

func (Pc *PostController) EditPost(c *fiber.Ctx) error {
	var req URLRequest

	claims := c.Locals("user").(*utils.Claims)
	// Parse the request body into the URLRequest struct
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
	}

	if err := Pc.validate.Struct(req); err != nil {
		// Return an error response if validation fails.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",               // Error message.
			"details": utils.FormatValidationErrors(err), // Formatting and sending detailed validation errors.
			"status":  fiber.StatusBadRequest,            // Unauthorized status code.
		})
	}

	postID := c.Params("id")
	existingPost, err := Pc.db.FindPostById(database.PostCreate{Id: postID})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to find post",
			"status": fiber.StatusInternalServerError,
		})
	}

	if existingPost.UserID != claims.UserID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":  "Unauthorized to edit this post",
			"status": fiber.StatusUnauthorized,
		})
	}

	updatedPost := database.PostCreate{
		Id:     existingPost.ID,
		Url:    html.EscapeString(req.URL),
		Title:  html.EscapeString(req.Title),
		Body:   html.EscapeString(req.Body),
		UserId: claims.UserID,
	}

	updatedPost, err := Pc.db.UpdatePost(updatedPost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update post",
			"status": fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Post updated successfully",
		"status":  fiber.StatusOK,
		"post": fiber.Map{
			"id":    updatedPost.Id,
			"User":  updatedPost.UserId,
			"Title": updatedPost.Title,
			"Body":  updatedPost.Body,
		},
	})
}

func (Pc *PostController) DeletePost(c *fiber.Ctx) error {
	postID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid post ID",
			"details": err.Error(),
		})
	}

	claims := c.Locals("user").(*utils.Claims)
	existingPost, err := Pc.db.FindPostById(database.PostCreate{Id: uint(postID)})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to find post",
			"status": fiber.StatusInternalServerError,
		})
	}

	if existingPost.UserID != claims.UserID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":  "Unauthorized to delete this post",
			"status": fiber.StatusUnauthorized,
		})
	}

	deletedPost, err := Pc.db.DeletePost(uint(postID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete post",
			"status": fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Post deleted successfully",
		"status":  fiber.StatusOK,
		"post": fiber.Map{
			"id":    deletedPost.ID,
			"User":  deletedPost.UserID,
			"Title": deletedPost.Title,
			"Body":  deletedPost.Body,
		},
	})
}
