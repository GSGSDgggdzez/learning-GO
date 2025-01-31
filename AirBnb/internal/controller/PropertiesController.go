package controllers

import (
	"AirBnb/internal/database"
	"AirBnb/internal/utils"
	"fmt"
	"html"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PropertiesController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

func NewPropertiesController(db database.Service) *PropertiesController {
	return &PropertiesController{
		db:       db,
		validate: validator.New(),
	}
}

// ----------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the registration logic -------------------------
// ----------------------------------------------------------------------------------------------------
type RegisterPropertyRequest struct {
	Title         string `form:"title" validate:"required,max=255"`
	Description   string `form:"description" validate:"required,max=255"`
	PricePerNight int    `form:"price_per_night" validate:"required,min=1"`
	Bedrooms      int    `form:"bed_room" validate:"required,min=1"`
	Guests        int    `form:"guests" validate:"required,min=1"`
	Country       string `form:"country" validate:"required,max=255"`
	CountryCode   string `form:"country_code" validate:"required,max=255"`
	Category      string `form:"category" validate:"required,max=255"`
}

func (pc *PropertiesController) RegisterProperty(c *fiber.Ctx) error {
	var req RegisterPropertyRequest

	// Extract user claims from the context
	claims := c.Locals("user").(*utils.Claims)

	// Parse the request body
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	// Validate the request struct
	if err := pc.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed, please check input", utils.FormatValidationErrors(err))
	}

	// Handle file upload
	file, err := c.FormFile("Image")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Profile picture is required", err.Error())
	}

	// Create a channel for the upload result
	uploadChan := make(chan struct {
		filePath string
		err      error
	})

	// Upload the file in a goroutine
	go func(file *multipart.FileHeader) {
		const maxFileSize = 10 * 1024 * 1024 // 10 MB

		var result struct {
			filePath string
			err      error
		}

		// Check file size
		if file.Size > maxFileSize {
			result.err = fmt.Errorf("file size exceeds the limit of 10 MB")
			uploadChan <- result
			return
		}

		// Create uploads directory if it doesn't exist
		uploadDir := "./image"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			if err := os.Mkdir(uploadDir, os.ModePerm); err != nil {
				result.err = err
				uploadChan <- result
				return
			}
		}

		// Save the file
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

	// Wait for the upload result
	uploadResult := <-uploadChan
	if uploadResult.err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to upload profile picture", uploadResult.err.Error())
	}

	// Prepare the property data for creation
	CreatePropertyData := database.Property{
		Title:         html.EscapeString(req.Title),
		Description:   html.EscapeString(req.Description),
		PricePerNight: req.PricePerNight, // No need to escape integers
		Bedrooms:      req.Bedrooms,      // No need to escape integers
		Guests:        req.Guests,        // No need to escape integers
		Country:       html.EscapeString(req.Country),
		CountryCode:   html.EscapeString(req.CountryCode),
		Category:      html.EscapeString(req.Category),
		Image:         uploadResult.filePath, // Use the uploaded file path
		LandlordID:    uint(claims.UserID),
	}

	// Create the property in the database
	newProperty, err := pc.db.CreateProperty(CreatePropertyData)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create property", err.Error())
	}

	// Return the created property
	// return c.Status(fiber.StatusCreated).JSON(newProperty)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Property registered successfully",
		"data": fiber.Map{
			"property": newProperty, // Ensure newProperty includes the ID
		},
	})
}

// ----------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the registration logic -------------------------
// ----------------------------------------------------------------------------------------------------

// ----------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the Delete logic -------------------------
// ----------------------------------------------------------------------------------------------------

func (pc *PropertiesController) DeleteProperty(c *fiber.Ctx) error {
	// Extract property ID from parameters
	propertyID := c.Params("id")

	// Convert the ID to an integer
	id, err := strconv.Atoi(propertyID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}

	// First get the property to have access to the image path
	property, err := pc.db.FindPropertyById(id)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to find property", err.Error())
	}

	if property == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Property not found", nil)
	}

	// Store the image path before deletion
	imagePath := property.Image

	// Delete the property from the database
	_, err = pc.db.DeleteProperty(uint(id))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete property", err.Error())
	}

	go func() {
		// Delete the associated image file
		if imagePath != "" {
			if err := os.Remove(imagePath); err != nil {
				fmt.Printf("Failed to delete image: %v\n", err)
			}
		}
	}()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Property successfully deleted",
		"property": fiber.Map{
			"property": property,
		},
	})
}

// ----------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the update logic -------------------------
// ----------------------------------------------------------------------------------------------------

type UpdatePropertyRequest struct {
	Title         string `form:"title" validate:"required,max=255"`
	Description   string `form:"description" validate:"required,max=255"`
	PricePerNight int    `form:"price_per_night" validate:"required,min=1"`
	Bedrooms      int    `form:"bed_room" validate:"required,min=1"`
	Guests        int    `form:"guests" validate:"required,min=1"`
	Country       string `form:"country" validate:"required,max=255"`
	CountryCode   string `form:"country_code" validate:"required,max=255"`
	Category      string `form:"category" validate:"required,max=255"`
}

func (pc *PropertiesController) UpdateProperty(c *fiber.Ctx) error {
	var req UpdatePropertyRequest

	propertyID := c.Params("id")

	// Extract user claims from the context
	claims := c.Locals("user").(*utils.Claims)

	// Parse the request body
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	// Validate the request struct
	if err := pc.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed, please check input", utils.FormatValidationErrors(err))
	}

	id, err := strconv.Atoi(propertyID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}

	// Find the property by ID
	FindProperty, err := pc.db.FindPropertyById(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Property not found", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update property", err.Error())
	}

	// Delete the old image in a goroutine
	go func() {
		if FindProperty.Image != "" {
			if err := os.Remove(FindProperty.Image); err != nil {
				fmt.Printf("Failed to delete image: %v\n", err)
			}
		}
	}()

	// Handle file upload
	file, err := c.FormFile("Image")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Profile picture is required", err.Error())
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid file type. Only JPEG, PNG, and GIF are allowed", nil)
	}

	// Upload the file
	uploadDir := "./image"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, os.ModePerm); err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create upload directory", err.Error())
		}
	}

	fileName := utils.GenerateUniqueFilename(file.Filename)
	filePath := filepath.Join(uploadDir, fileName)
	if err := c.SaveFile(file, filePath); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to save uploaded file", err.Error())
	}

	// Prepare the update data
	UpdatePropertyData := database.Property{
		Title:         html.EscapeString(req.Title),
		Description:   html.EscapeString(req.Description),
		PricePerNight: req.PricePerNight,
		Bedrooms:      req.Bedrooms,
		Guests:        req.Guests,
		Country:       html.EscapeString(req.Country),
		CountryCode:   html.EscapeString(req.CountryCode),
		Category:      html.EscapeString(req.Category),
		Image:         filePath,
		LandlordID:    uint(claims.UserID),
	}

	UpdateProperty, err := pc.db.UpdateProperty(uint(id), UpdatePropertyData)

	// Update the property in the database
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update property", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Property registered successfully",
		"data": fiber.Map{
			"user": fiber.Map{
				"ID":   claims.ID,
				"Name": claims.Name, // Sanitize user input
			},
			"property": UpdateProperty, // Ensure newProperty includes the ID
		},
	})
}

// ----------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the Find ALL logic -------------------------
// ----------------------------------------------------------------------------------------------------

func (pc *PropertiesController) GetAllProperties(c *fiber.Ctx) error {
	// Get pagination parameters from query
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "0")) // 0 means no limit

	properties, total, err := pc.db.FindAllProperties(page, limit)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch properties", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": fiber.StatusOK,
		"data":   properties,
		"meta": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// ----------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the Find By ID logic -------------------------
// ----------------------------------------------------------------------------------------------------

func (pc *PropertiesController) GetPropertyById(c *fiber.Ctx) error {
	propertyID, err := c.ParamsInt("id")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}

	property, err := pc.db.FindPropertyById(propertyID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch property", err.Error())
	}

	if property == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Property not found", nil)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": fiber.StatusOK,
		"data":   property,
	})
}
