package controllers

import (
	"Tiktok/internal/database"
	"Tiktok/internal/utils"
	"html"
	"strconv"
	"strings"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type CommentController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

func NewCommentController(db database.Service) *CommentController {
	return &CommentController{
		db:       db,              // Setting the provided database service.
		validate: validator.New(), // Initializing a new validator instance.
	}
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the Start of Create post Comment logic -------------------------
// --------------------------------------------------------------------------------------------------

type CommentPostRequest struct {
	Comment string `json:"comment" validate:"required, max=255"`
}

func (cc *CommentController) CommentPost(c *fiber.Ctx) error {
	postID, err := c.ParamsInt("id")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid post ID", err.Error())
	}

	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or missing authentication", "")
	}

	var req CommentPostRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request data", err.Error())
	}

	if err := cc.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	Comment := database.Comment{
		UserID: uint(claims.UserID),
		PostID: uint(postID),
		Text:   html.EscapeString(strings.TrimSpace(req.Comment)),
	}

	CreateComment, err := cc.db.CreateComment(Comment)

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update post", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Post updated successfully",
		"post":    CreateComment,
	})
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of Create comment logic -------------------------
// --------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of Update comment logic -------------------------
// --------------------------------------------------------------------------------------------------

type UpdateCommentPostRequest struct {
	Comment string `json:"comment" validate:"required, max=255"`
}

func (cc *CommentController) UpdateCommentPost(c *fiber.Ctx) error {

	postID, err := c.ParamsInt("id")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid post ID", err.Error())
	}

	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or missing authentication", "")
	}

	var req UpdateCommentPostRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request data", err.Error())
	}

	if err := cc.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	existingPost, err := cc.db.FindCommentById(uint(postID))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Post not found", err.Error())
	}

	if existingPost.UserID != uint(claims.UserID) {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Not authorized to edit this post", "")
	}

	if req.Comment != "" {
		existingPost.Text = html.EscapeString(req.Comment)
	}
	result, err := cc.db.UpdateComment(*existingPost)

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update post", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Post updated successfully",
		"post":    result,
	})
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of update comment logic -------------------------
// --------------------------------------------------------------------------------------------------

func (cc *CommentController) DeleteComment(c *fiber.Ctx) error {
	commentID := c.Params("id")

	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or missing authentication", "")
	}

	claimsID := strconv.Itoa(claims.UserID)

	commentIDUint, err := strconv.ParseUint(commentID, 10, 32)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid comment ID", "Comment ID must be a positive number")
	}

	findComment, err := cc.db.FindCommentById(uint(commentIDUint))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Comment not found", err.Error())
	}

	if claimsID != strconv.FormatUint(uint64(findComment.UserID), 10) {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "You are not authorized to delete this comment", "")
	}

	err = cc.db.DeleteComment(uint(commentIDUint))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete comment", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Comment deleted successfully",
	})
}
