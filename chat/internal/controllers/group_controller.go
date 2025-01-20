package controllers

import (
	"chat/internal/database"
	"chat/internal/utils"
	"fmt"
	"html"
	"mime/multipart"
	"os"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type GroupController struct {
	db       database.Service
	validate *validator.Validate
}

func NewGroupController(db database.Service) *GroupController {
	return &GroupController{
		db:       db,
		validate: validator.New(),
	}
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the Create Groupe logic -------------------------
// ---------------------------------------------------------------------------------------------------

type CreateGroupRequest struct {
	Name        string `form:"name" validate:"required,max=255"`
	Description string `form:"description" validate:"required,max=255" `
}

func (gc *GroupController) Create(c *fiber.Ctx) error {
	var req CreateGroupRequest
	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
	}

	if err := gc.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": utils.FormatValidationErrors(err),
			"status":  fiber.StatusBadRequest,
		})
	}

	var filePath string
	file, err := c.FormFile("avatar")
	if err == nil && file != nil {
		uploadChan := make(chan struct {
			filePath string
			err      error
		})

		go func(file *multipart.FileHeader) {
			const maxFileSize = 10 * 1024 * 1024
			var result struct {
				filePath string
				err      error
			}

			if file.Size > maxFileSize {
				result.err = fmt.Errorf("file size exceeds maximum allowed")
				uploadChan <- result
				return
			}

			uploadDir := "./GroupUploads"
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				if err := os.Mkdir(uploadDir, os.ModePerm); err != nil {
					result.err = err
					uploadChan <- result
					return
				}
			}

			fileName := utils.GenerateUniqueFilename(file.Filename)
			filePath := fmt.Sprintf("%s/%s", uploadDir, fileName)
			if err := c.SaveFile(file, filePath); err != nil {
				result.err = err
				uploadChan <- result
				return
			}

			result.filePath = filePath
			uploadChan <- result
		}(file)

		uploadResult := <-uploadChan
		if uploadResult.err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload profile picture",
			})
		}
		filePath = uploadResult.filePath
	}

	CreateGroup := database.GroupData{
		Name:        html.EscapeString(req.Name),
		Description: html.EscapeString(req.Description),
		Image:       filePath,
		OwnerId:     claims.UserID,
	}

	NewGroup, err := gc.db.CreateGroup(CreateGroup)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create NewGroup",
			"status": fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Group created successfully",
		"status":  fiber.StatusCreated,
		"Group": fiber.Map{
			"ID":          NewGroup.Id,
			"Name":        NewGroup.Name,
			"Description": NewGroup.Description,
			"Owner_Id":    NewGroup.OwnerId,
			"Image":       NewGroup.Image,
		},
	})
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of the Create Groupe logic -------------------------
// ---------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the Find Groupe logic -------------------------
// ---------------------------------------------------------------------------------------------------

func (gc *GroupController) FindGroup(c *fiber.Ctx) error {
	id := c.Params("id")

	groupId, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid group ID",
			"status": fiber.StatusBadRequest,
		})
	}

	// Change to use correct function
	group, err := gc.db.FindGroupById(groupId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Error finding group",
			"status": fiber.StatusInternalServerError,
		})
	}

	if group == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":  "Group not found",
			"status": fiber.StatusNotFound,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"group":  group,
		"status": fiber.StatusOK,
	})
}

func (gc *GroupController) GetAllGroups(c *fiber.Ctx) error {
	// Get optional pagination params
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	// Get groups from database
	groups, total, err := gc.db.FindAllGroups(page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to fetch groups",
			"status": fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"groups": groups,
		"total":  total,
		"page":   page,
		"limit":  limit,
		"status": fiber.StatusOK,
	})
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of the Find Groupe logic -------------------------
// ---------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the Delete Groupe logic -------------------------
// ---------------------------------------------------------------------------------------------------

func (gc *GroupController) DeleteGroup(c *fiber.Ctx) error {
	// Get ID from params instead of query
	id := c.Params("id")

	groupId, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid group ID",
			"status": fiber.StatusBadRequest,
		})
	}

	// Get claims
	claims := c.Locals("user").(*utils.Claims)

	// Find group first
	group, err := gc.db.FindGroupById(groupId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Error finding group",
			"status": fiber.StatusInternalServerError,
		})
	}

	if group == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":  "Group not found",
			"status": fiber.StatusNotFound,
		})
	}

	// Check ownership
	if claims.UserID != group.OwnerId {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":  "Not authorized to delete this group",
			"status": fiber.StatusUnauthorized,
		})
	}

	// Delete group messages first
	if err := gc.db.DeleteGroupMessages(groupId); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete group messages",
			"status": fiber.StatusInternalServerError,
		})
	}

	// Delete group
	if err := gc.db.DeleteGroup(groupId); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete group",
			"status": fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Group deleted successfully",
		"status":  fiber.StatusOK,
	})
}
