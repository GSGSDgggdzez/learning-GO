package controllers

import (
	"JobBoard/internal/database"
	"JobBoard/utils"
	"fmt"
	"html"
	"mime/multipart"
	"os"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type CompanyController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

func NewCompanyController(db database.Service) *CompanyController {
	return &CompanyController{
		db:       db,              // Setting the provided database service.
		validate: validator.New(), // Initializing a new validator instance.
	}
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the registration logic -------------------------
// ---------------------------------------------------------------------------------------------------

type RegisterCompanyRequest struct {
	Name        string `form:"name" validate:"required,max=255"`
	Description string `form:"description" validate:"required"`
	Website     string `form:"website" validate:"required,max=255"`
	Location    string `form:"location" validate:"required,max=255"`
	Email       string `form:"email" validate:"required,email,max=255"`
}

func (cc *CompanyController) RegisterCompany(c *fiber.Ctx) error {
	var req RegisterCompanyRequest // Declare a variable of type RegisterRequest to bind form data.

	claims, ok := c.Locals("user").(*utils.Claims)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
	}

	// Step 2: Validate the request data
	// Validate the fields in RegisterRequest using the 'validate' instance.
	if err := cc.validate.Struct(req); err != nil {
		// Return an error response if validation fails.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",               // Error message.
			"details": utils.FormatValidationErrors(err), // Formatting and sending detailed validation errors.
			"status":  fiber.StatusBadRequest,            // Unauthorized status code.
		})
	}

	// Create channel for file upload result
	uploadChan := make(chan struct {
		filePath string
		err      error
	})

	// Step 7: Handle profile picture upload
	// Handle profile picture upload in goroutine
	file, err := c.FormFile("profile_picture")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Profile picture is required",
		})
	}

	go func(file *multipart.FileHeader) {

		const maxFileSize = 10 * 1024 * 1024

		var result struct {
			filePath string
			err      error
		}

		if file.Size > maxFileSize {
			result.err = err
			return
		}

		// Create uploads directory
		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			if err := os.Mkdir(uploadDir, os.ModePerm); err != nil {
				result.err = err
				uploadChan <- result
				return
			}
		}

		// Save file
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
			"error": "Failed to upload company logo",
		})
	}

	existingCompany, err := cc.db.FindCompanyByEmail(req.Email)
	if err == nil && existingCompany != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "A company with this email already exists",
			"status": fiber.StatusBadRequest,
		})
	}

	// Step 8: Prepare user data for creation
	// Escape the user's email and name to prevent potential XSS vulnerabilities.
	createCompanyDate := database.CompanyCreate{
		Email:       html.EscapeString(req.Email), // Safely escape the email.     // Store the hashed password.
		Name:        html.EscapeString(req.Name),  // Safely escape the name.
		Logo:        uploadResult.filePath,
		Description: html.EscapeString(req.Description),
		Location:    html.EscapeString(req.Location),
		Website:     html.EscapeString(req.Website),
	}

	// Step 9: Create the user in the database
	// Attempt to create a new user in the database.
	newCompany, err := cc.db.CreateCompany(createCompanyDate)
	if err != nil {
		// Return an error response if user creation fails.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create company",      // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
			"detail": err.Error(),
		})
	}

	// first find the user

	// update the User table to add the company Id to employer

	if err := cc.db.UpdateUserCompanyID(int(claims.UserID), newCompany.ID, database.UserCreate{}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update user with company ID",
			"status": fiber.StatusInternalServerError,
			"detail": err.Error(),
		})
	}

	user, err := cc.db.FindUserByEmail(claims.Email)

	if user != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":  "Invalid email or password",
			"status": fiber.StatusUnauthorized,
		})
	}

	// Step 11: Respond with success
	// Return a success response, asking the user to verify their email.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"detail": "User have been verify successfully",
		"status": fiber.StatusOK,
		"User": fiber.Map{
			"id":         user.ID,
			"email":      user.Email,
			"Name":       user.Name,
			"profileUrl": user.Avatar,
			"IsEmployer": user.IsEmployer,
			"companyID":  user.CompanyID,
			"Company": fiber.Map{
				"id":          newCompany.ID,
				"email":       newCompany.Email,
				"Name":        newCompany.Name,
				"profileUrl":  newCompany.Logo,
				"Location":    newCompany.Location,
				"Website":     newCompany.Website,
				"Description": newCompany.Description,
			},
		},
	})

}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the end of the registration logic -------------------------
// --------------------------------------------------------------------------------------------------
