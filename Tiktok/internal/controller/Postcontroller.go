package controllers

import (
	"Tiktok/internal/config"
	"Tiktok/internal/database"
	"Tiktok/internal/utils"
	"context"
	"html"
	"mime/multipart"
	"strings"
	"time"

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

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the create post logic -------------------------
// --------------------------------------------------------------------------------------------------

type CreatePostRequest struct {
	Text      string `form:"text" validate:"required,max=255,min=1"`
	Hashtags  string `form:"hashtags" validate:"required"`
	Music     string `form:"music" validate:"required,max=255"`
	Location  string `form:"location" validate:"required,max=255"`
	IsPrivate bool   `form:"is_private"`
}

func (pc *PostController) CreatePost(c *fiber.Ctx) error {
	var req CreatePostRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := pc.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":  "Invalid or missing authentication",
			"status": fiber.StatusUnauthorized,
		})
	}

	file, err := c.FormFile("video")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Video file is required", err.Error())
	}

	if file.Size > 100*1024*1024 { // 100MB limit
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Video file too large", "Maximum file size is 100MB")
	}

	uploadChan := make(chan struct {
		url      string
		duration float64
		err      error
	})

	cld, err := config.InitCloudinary()
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to initialize Cloudinary", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	go func(file *multipart.FileHeader) {
		var result struct {
			url      string
			duration float64
			err      error
		}

		if err := utils.ValidateVideoFile(file); err != nil {
			result.err = err
			uploadChan <- result
			return
		}

		fileHeader, err := file.Open()
		if err != nil {
			result.err = err
			uploadChan <- result
			return
		}
		defer fileHeader.Close()

		url, err := utils.UploadToCloudinary(cld, ctx, fileHeader)
		if err != nil {
			result.err = err
			uploadChan <- result
			return
		}

		result.url = url
		uploadChan <- result
	}(file)

	select {
	case result := <-uploadChan:
		if result.err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Upload failed", result.err.Error())
		}

		sanitizedText := html.EscapeString(strings.TrimSpace(req.Text))
		sanitizedLocation := html.EscapeString(strings.TrimSpace(req.Location))
		sanitizedMusic := html.EscapeString(strings.TrimSpace(req.Music))

		hashtagSlice := strings.Split(req.Hashtags, ",")
		var cleanedHashtags []string
		for _, tag := range hashtagSlice {
			tag = strings.TrimSpace(tag)
			tag = strings.TrimPrefix(tag, "#")
			if tag != "" {
				cleanedHashtags = append(cleanedHashtags, tag)
			}
		}

		post := database.Post{
			UserID:    database.User{ID: uint(claims.UserID)},
			Text:      sanitizedText,
			Video:     result.url,
			IsPrivate: req.IsPrivate,
			Music:     sanitizedMusic,
			Location:  sanitizedLocation,
			Duration:  result.duration,
		}

		createdPost, err := pc.db.CreatePost(post, cleanedHashtags)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to save post", err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Post created successfully",
			"post":    createdPost,
		})

	case <-ctx.Done():
		return utils.SendErrorResponse(c, fiber.StatusGatewayTimeout, "Upload timeout", "Request took too long to process")
	}
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of the create Post logic -------------------------
// --------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the update post logic -------------------------
// --------------------------------------------------------------------------------------------------

type UpdatePostRequest struct {
	Text      string `json:"text" validate:"required,max=255,min=1"`
	Hashtags  string `json:"hashtags" validate:"required"`
	Music     string `json:"music" validate:"required,max=255"`
	Location  string `json:"location" validate:"required,max=255"`
	IsPrivate bool   `json:"is_private"`
}

func (pc *PostController) UpdatePost(c *fiber.Ctx) error {
	postID, err := c.ParamsInt("id")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid post ID", err.Error())
	}

	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or missing authentication", "")
	}

	var req UpdatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request data", err.Error())
	}

	if err := pc.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	existingPost, err := pc.db.FindPostById(uint(postID))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Post not found", err.Error())
	}

	if existingPost.UserID != uint(claims.UserID) {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Not authorized to edit this post", "")
	}

	hashtagSlice := strings.Split(req.Hashtags, ",")
	var cleanedHashtags []string
	for _, tag := range hashtagSlice {
		tag = strings.TrimSpace(tag)
		tag = strings.TrimPrefix(tag, "#")
		if tag != "" {
			cleanedHashtags = append(cleanedHashtags, tag)
		}
	}

	updatedPost := database.Post{
		ID:        uint(postID),
		Text:      html.EscapeString(strings.TrimSpace(req.Text)),
		Music:     html.EscapeString(strings.TrimSpace(req.Music)),
		Location:  html.EscapeString(strings.TrimSpace(req.Location)),
		IsPrivate: req.IsPrivate,
		Video:     existingPost.Video,
		Duration:  existingPost.Duration,
		UserID:    database.User{ID: existingPost.User.ID},
	}

	result, err := pc.db.UpdatePost(updatedPost, cleanedHashtags)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update post", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Post updated successfully",
		"post":    result,
	})
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of Update logic -------------------------
// --------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the delete logic -------------------------
// --------------------------------------------------------------------------------------------------

func (pc *PostController) DeletePost(c *fiber.Ctx) error {
	postID, err := c.ParamsInt("id")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid post ID", err.Error())
	}

	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or missing authentication", "")
	}

	existingPost, err := pc.db.FindPostById(uint(postID))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Post not found", err.Error())
	}

	if existingPost.UserID != uint(claims.UserID) {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Not authorized to delete this post", "")
	}

	if err := pc.db.DeletePost(uint(postID)); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete post", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Post deleted successfully",
	})
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of the delete logic -------------------------
// --------------------------------------------------------------------------------------------------
