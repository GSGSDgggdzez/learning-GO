package controllers

import (
	"Video_to_text/internal/database" // Importing the database service for accessing DB-related functionalities.
	"Video_to_text/internal/utils"    // Importing utility functions like token generation and email handling.
	"fmt"
	"html" // Importing the 'html' package to escape strings for security.
	"mime/multipart"
	"os"
	"strings"

	"github.com/go-playground/validator" // Importing validator package for input validation.
	"github.com/gofiber/fiber/v2"        // Importing Fiber framework for web handling.
	"golang.org/x/crypto/bcrypt"         // Importing bcrypt for secure password hashing.
)

// AuthController handles authentication-related actions like user registration.
type AuthController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

// NewAuthController creates a new instance of AuthController with a database service.
func NewAuthController(db database.Service) *AuthController {
	return &AuthController{
		db:       db,              // Setting the provided database service.
		validate: validator.New(), // Initializing a new validator instance.
	}
}

type RegisterRequest struct {
	Email    string `form:"email" validate:"required,email,max=255"`    // Validates the email field.
	Password string `form:"password" validate:"required,min=8,max=255"` // Validates password (minimum length 8).
	Name     string `form:"name" validate:"required,max=255"`           // Validates that name is provided.			   // Validates that a profile image is provided.
}

// Register handles the user registration process.
func (ac *AuthController) Register(c *fiber.Ctx) error {
	// Step 1: Parse the incoming multipart form
	var req RegisterRequest // Declare a variable of type RegisterRequest to bind form data.

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
	}

	// Step 2: Validate the request data
	// Validate the fields in RegisterRequest using the 'validate' instance.
	if err := ac.validate.Struct(req); err != nil {
		// Return an error response if validation fails.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",               // Error message.
			"details": utils.FormatValidationErrors(err), // Formatting and sending detailed validation errors.
			"status":  fiber.StatusBadRequest,            // Unauthorized status code.
		})
	}

	// Step 3: Check if the user already exists
	// Attempt to find an existing user by email.
	findUser, err := ac.db.FindUserByEmail(req.Email, req.Password)
	if err != nil {
		// Return an error response if the user is not found.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "User Not found",       // Error message.
			"detail": err.Error(),            // Error details.
			"status": "fiber.StatusNotFound", // Status code for not found.
		})
	}

	// Step 4: Handle user existence check
	// If a user already exists with the same email, return an error message.
	if findUser != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "User already registered, please login", // Error message.
			"status": fiber.StatusBadRequest,                  // Bad request status code.
		})
	}

	// Step 5: Hash the password
	// Securely hash the user's password using bcrypt.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		// Return an error response if password hashing fails.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to process password", // Error message.
			"status": fiber.StatusUnauthorized,     // Unauthorized status code.
		})
	}

	// Step 6: Generate a verification token
	// Generate a unique token to verify the user's email.
	token, err := utils.GenerateVerificationToken()
	if err != nil {
		// Return an error response if token generation fails.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": fiber.StatusInternalServerError, // Internal server error status.
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
			"error": "Failed to upload profile picture",
		})
	}

	// Step 8: Prepare user data for creation
	// Escape the user's email and name to prevent potential XSS vulnerabilities.
	createUserDate := database.UserCreate{
		Email:       html.EscapeString(req.Email), // Safely escape the email.
		Password:    string(hashedPassword),       // Store the hashed password.
		Name:        html.EscapeString(req.Name),  // Safely escape the name.
		Token:       token,                        // Attach the verification token.
		Profile_Url: uploadResult.filePath,        // Store the profile picture path.
	}

	// Step 9: Create the user in the database
	// Attempt to create a new user in the database.
	newUser, err := ac.db.CreateUser(createUserDate)
	if err != nil {
		// Return an error response if user creation fails.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create user",         // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
		})
	}

	// Step 10: Send verification email
	// After user creation, send a verification email with the token.
	// Send the verification email asynchronously using a goroutine
	go func() {
		// Assume you have a function that sends the email
		err := utils.SendVerificationEmail(newUser.Email, newUser.Token)
		if err != nil {
			// Return an error response if sending the verification email fails.
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to send verification email", // Error message.
				"status": fiber.StatusInternalServerError,     // Internal server error status code.
				"TO":     newUser.Email,
			})
		}
	}()

	// Step 11: Respond with success
	// Return a success response, asking the user to verify their email.
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully, please check your email for verification", // Success message.
		"status":  fiber.StatusCreated,                                                   // Created status code.
	})
}

func (ac *AuthController) VerifyEmail(c *fiber.Ctx) error {
	token := strings.TrimSpace(html.EscapeString(c.Query("token")))
	if len(token) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Verification token is required",
		})
	}

	updateResult, err := ac.db.VerifyUserAndUpdate(token)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create user",         // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
			"Error":  err.Error(),
		})
	}

	JWT, err := utils.GenerateToken(updateResult.ID, updateResult.Profile_Url, updateResult.Email, updateResult.Name, updateResult.Token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to generate token",      // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{

		"detail": "User have been verify successfully",
		"status": fiber.StatusOK,
		"JWT":    JWT,
		"user": fiber.Map{
			"id":         updateResult.ID,
			"email":      updateResult.Email,
			"Name":       updateResult.Name,
			"profileUrl": updateResult.Profile_Url,
		},
	})

}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"` // Must be a valid email
	Password string `json:"password" validate:"required"`    // Cannot be empty
}

func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	user, err := ac.db.FindUserByEmail(req.Email, req.Password)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Error during login",
			"status": fiber.StatusInternalServerError,
		})
	}

	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":  "Invalid email or password",
			"status": fiber.StatusUnauthorized,
		})
	}

	JWT, err := utils.GenerateToken(user.ID, user.Profile_Url, user.Email, user.Name, user.Token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to generate token",      // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{

		"detail": "User have been verify successfully",
		"status": fiber.StatusOK,
		"JWT":    JWT,
		"user": fiber.Map{
			"id":         user.ID,
			"email":      user.Email,
			"Name":       user.Name,
			"profileUrl": user.Profile_Url,
		},
	})

}

func (ac *AuthController) DeleteUser(c *fiber.Ctx) error {
	// Get claims from context
	claims := c.Locals("user").(*utils.Claims)
	DeleteUser, err := ac.db.DeleteUser(claims.UserID)

	if DeleteUser.Profile_Url != "" {
		if err := os.Remove(DeleteUser.Profile_Url); err != nil {
			// Log the error but continue with user deletion
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to delete user profile picture", // Error message.
				"status": fiber.StatusInternalServerError,         // Internal server error status code.
			})
		}
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete user ",        // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":  "User deleted successfully", // Error message.
		"status": fiber.StatusOK,              // Internal server error status code.
	})

}

type UpdateRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Name     string `json:"name" validate:"required,max=255"`
	Password string `json:"password" validate:"required,min=8,max=255"`
}

func (ac *AuthController) UpdateUser(c *fiber.Ctx) error {
	// Get claims from context
	claims := c.Locals("user").(*utils.Claims)

	// Parse request body
	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := ac.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": utils.FormatValidationErrors(err),
		})
	}

	file, err := c.FormFile("profile_picture")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Profile picture is required",
		})
	}

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
			"error": "Failed to upload profile picture",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		// Return an error response if password hashing fails.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to process password", // Error message.
			"status": fiber.StatusUnauthorized,     // Unauthorized status code.
		})
	}

	createUserDate := database.UserCreate{
		Email:       html.EscapeString(req.Email), // Safely escape the email.
		Password:    string(hashedPassword),       // Store the hashed password.
		Name:        html.EscapeString(req.Name),  // Safely escape the name.
		Profile_Url: uploadResult.filePath,        // Store the profile picture path.
	} // Add the missing closing brace
	// Update controller code
	// Update controller code
	updatedUser, err := ac.db.UpdateUser(claims.UserID, database.UserUpdate{
		Email:       createUserDate.Email,
		Password:    createUserDate.Password,
		Name:        createUserDate.Name,
		Profile_Url: createUserDate.Profile_Url,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update user email",   // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
		})
	}

	JWT, err := utils.GenerateToken(updatedUser.ID, updatedUser.Profile_Url, updatedUser.Email, updatedUser.Name, updatedUser.Token)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update user email",   // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Email updated successfully",
		"NewJWT":  JWT,
		"user": fiber.Map{
			"id":         updatedUser.ID,
			"email":      updatedUser.Email,
			"Name":       updatedUser.Name,
			"profileUrl": updatedUser.Profile_Url,
		},
	})
}

func (ac *AuthController) ForgotPassword(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing or invalid Authorization header",
		})
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token and get claims
	claims, err := utils.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	err = utils.SendVerificationPassword(claims.Email, claims.JWTToken)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to send email",
			"detail": err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	// TODO: Implement password reset logic here using the claims data

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password reset instructions sent to email",
	})
}

func (ac *AuthController) ResetPassword(c *fiber.Ctx) error {
	token := strings.TrimSpace(html.EscapeString(c.Query("token")))
	if len(token) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Verification token is required",
		})
	}

	User, err := ac.db.FindUserByToken(token)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Failed to verify email",
			"detail": err.Error(),
		})
	}

	JWT, err := utils.GenerateToken(User.ID, User.Profile_Url, User.Email, User.Name, User.Token)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Email updated successfully",
		"NewJWT":  JWT,
		"user": fiber.Map{
			"id":         User.ID,
			"email":      User.Email,
			"Name":       User.Name,
			"profileUrl": User.Profile_Url,
		},
	})
}
