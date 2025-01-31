package controllers

import (
	"AirBnb/internal/database"
	"AirBnb/internal/utils"
	"fmt"
	"html"
	"mime/multipart"
	"os"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the registration logic -------------------------
// ---------------------------------------------------------------------------------------------------

type RegisterRequest struct {
	Name     string `form:"name" validate:"required,max=255"`
	Email    string `form:"email" validate:"required,email,max=255"`
	Password string `form:"password" validate:"required,max=255"`
}

func (ac *AuthController) Register(c *fiber.Ctx) error {
	var req RegisterRequest

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

	findUser, err := ac.db.FindUserByEmail(req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		// Return an error response if there is an error other than record not found.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to check user existence", // Error message.
			"detail": err.Error(),                      // Error details.
			"status": fiber.StatusInternalServerError,  // Internal server error status code.
		})
	}

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
	file, err := c.FormFile("avatar")
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
		uploadDir := "./image"
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
	createUserDate := database.User{
		Email:    html.EscapeString(req.Email), // Safely escape the email.
		Password: string(hashedPassword),       // Store the hashed password.
		Name:     html.EscapeString(req.Name),  // Safely escape the name.
		Token:    token,                        // Attach the verification token.
		Avatar:   uploadResult.filePath,
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
		"status":  fiber.StatusCreated,
		"userID":  newUser.Email, // Created status code.
	})

}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the end of the registration logic -------------------------
// --------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the Verify Email logic -------------------------
// --------------------------------------------------------------------------------------------------

func (ac *AuthController) VerifyEmail(c *fiber.Ctx) error {
	token := c.Params("token")

	updateResult, err := ac.db.VerifyUserAndUpdate(token)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create user",         // Error message.
			"status": fiber.StatusInternalServerError, // Internal server error status code.
			"Error":  err.Error(),
		})
	}

	JWT, err := utils.GenerateToken(int(updateResult.ID), updateResult.Avatar, updateResult.Email, updateResult.Name, updateResult.Token, updateResult.IsActive, updateResult.IsStaff, updateResult.IsVerified)
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
			"profileUrl": updateResult.Avatar,
			"IsVerified": updateResult.IsVerified,
			"IsStaff":    updateResult.IsStaff,
			"IsActive":   updateResult.IsActive,
		},
	})

}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of the Verify Email logic -------------------------
// --------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the Start of the LoginRequest logic -------------------------
// --------------------------------------------------------------------------------------------------

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

	user, err := ac.db.FindUserByEmail(req.Email)

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

	JWT, err := utils.GenerateToken(int(user.ID), user.Avatar, user.Email, user.Name, user.Token, user.IsActive, user.IsStaff, user.IsVerified)
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
			"profileUrl": user.Avatar,
			"IsStaff":    user.IsStaff,
			"IsActive":   user.IsActive,
			"IsVerified": user.IsVerified,
		},
	})

}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of the LoginRequest logic -------------------------
// --------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the ForgotPasswordRequest logic -------------------------
// --------------------------------------------------------------------------------------------------
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=255"` // Must be a valid email
}

func (ac *AuthController) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest

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

	findUser, err := ac.db.FindUserByEmail(req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		// Return an error response if there is an error other than record not found.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to check user existence", // Error message.
			"detail": err.Error(),                      // Error details.
			"status": fiber.StatusInternalServerError,  // Internal server error status code.
		})
	}

	go func() {
		// Assume you have a function that sends the email
		err := utils.SendVerificationPassword(findUser.Email, findUser.Token)
		if err != nil {
			// Return an error response if sending the verification email fails.
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to send verification email", // Error message.
				"status": fiber.StatusInternalServerError,     // Internal server error status code.
				"TO":     findUser.Email,
			})
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Reset Password Email have been send please check you email", // Success message.
		"status":  fiber.StatusCreated,                                          // Created status code.
	})

}

// -----------------------------------------------------------------------------------------------------
// ------------------------------ these is the End of the ForgotPasswordRequest logic -------------------------
// --------------------------------------------------------------------------------------------------

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the start  of the RestPasswordRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

func (ac *AuthController) RestPassword(c *fiber.Ctx) error {
	Token := c.Params("Token")

	FindToken, err := ac.db.FindUserByToken(Token)

	if err != nil && err != gorm.ErrRecordNotFound {
		// Return an error response if there is an error other than record not found.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to check user existence", // Error message.
			"detail": err.Error(),                      // Error details.
			"status": fiber.StatusInternalServerError,  // Internal server error status code.
		})
	}

	if FindToken == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":  "User not found", // Change here
			"status": fiber.StatusNotFound,
		})
	}

	JWT, err := utils.GenerateToken(int(FindToken.ID), FindToken.Avatar, FindToken.Email, FindToken.Name, FindToken.Token, FindToken.IsActive, FindToken.IsStaff, FindToken.IsVerified)
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
			"id":         FindToken.ID,
			"email":      FindToken.Email,
			"Name":       FindToken.Name,
			"profileUrl": FindToken.Avatar,
			"IsStaff":    FindToken.IsStaff,
			"IsActive":   FindToken.IsActive,
			"IsVerified": FindToken.IsVerified,
		},
	})

}

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the End of the RestPasswordRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the Start of the DeleteRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

func (ac *AuthController) DeleteUser(c *fiber.Ctx) error {
	// Extract user claims from the context
	claims := c.Locals("user").(*utils.Claims)

	// Extract user ID from parameters
	ID := c.Params("ID")

	claimsID := strconv.Itoa(claims.UserID)

	// Check if the user is authorized to delete the account
	if claimsID != ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "You are not authorized to delete this account", // Error message.
			"status": fiber.StatusBadRequest,                          // Bad request status code.
		})
	}

	// Find the user by email to retrieve their details (e.g., avatar)
	deleteUser, err := ac.db.DeleteUser(ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		// Handle errors other than record not found
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete user", // Error message.
			"detail": err.Error(),             // Error details.
			"status": fiber.StatusInternalServerError,
		})
	}

	// If the user does not exist
	if deleteUser == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":  "User not found", // Error message.
			"status": fiber.StatusNotFound,
		})
	}

	// Delete the user's profile picture from the ./image directory
	if deleteUser.Avatar != "" {
		err := os.Remove(deleteUser.Avatar)
		if err != nil {
			// Log the error but do not block the request
			fmt.Printf("Failed to delete image: %s, error: %v\n", deleteUser.Avatar, err)
		}
	}

	// Return a success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User successfully deleted",
		"user": fiber.Map{
			"id":    deleteUser.ID,
			"email": deleteUser.Email,
			"name":  deleteUser.Name,
		},
	})
}

// TODO: Updata User data
