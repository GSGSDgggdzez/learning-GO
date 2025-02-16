package controllers

import (
	"Tiktok/internal/config"
	"Tiktok/internal/database"
	"Tiktok/internal/utils"
	"context"
	"errors"
	"html"
	"log"
	"mime/multipart"
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
	Password string `form:"password" validate:"required,max=255,min=8"`
	Bio      string `form:"bio" validate:"max=255"`
}

func (ac *AuthController) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := ac.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Profile picture is required", err.Error())
	}

	// Create upload channel
	uploadChan := make(chan struct {
		url string
		err error
	})

	// Get Cloudinary instance before goroutine
	cld, err := config.InitCloudinary()
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to initialize Cloudinary", err.Error())
	}

	// Create context
	ctx := context.Background()

	// Handle upload in goroutine
	go func(file *multipart.FileHeader) {
		var result struct {
			url string
			err error
		}

		if err := utils.ValidateImageFile(file); err != nil {
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

	token, err := utils.GenerateVerificationToken()
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate verification token", err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
	}

	// Wait for upload result
	uploadResult := <-uploadChan
	if uploadResult.err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to upload image", uploadResult.err.Error())
	}

	createUserData := database.User{
		Name:     html.EscapeString(req.Name),
		Email:    html.EscapeString(req.Email),
		Password: string(hashedPassword),
		Bio:      html.EscapeString(req.Bio),
		Avatar:   uploadResult.url,
		Token:    token,
	}

	NewUser, err := ac.db.CreateUser(createUserData)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user", err.Error())
	}

	go func() {
		err := utils.SendVerificationEmail(NewUser.Email, NewUser.Token)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to send verification email",
				"status": fiber.StatusInternalServerError,
				"TO":     NewUser.Email,
			})
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully, please check your email for verification",
		"status":  fiber.StatusCreated,
		"userID":  NewUser.Email,
	})
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the Verify Email logic -------------------------
// --------------------------------------------------------------------------------------------------

func (ac *AuthController) VerifyEmail(c *fiber.Ctx) error {
	Badtoken := c.Params("token")

	GodTOken := html.EscapeString(Badtoken)

	updateResult, err := ac.db.VerifyUserAndUpdate(GodTOken)

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Bad Verification token", err.Error())
	}

	JWT, err := utils.GenerateToken(int(updateResult.ID), updateResult.Avatar, updateResult.Email, updateResult.Name, updateResult.Token, updateResult.Bio, updateResult.EmailVerified, updateResult.FollowerCount, updateResult.FollowingCount)

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate JWT token", err.Error())
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
			"IsVerified": updateResult.EmailVerified,
			"Bio":        updateResult.Bio,
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
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := ac.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	user, err := ac.db.FindUserByEmail(html.EscapeString(req.Email))

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Error during login", utils.FormatValidationErrors(err))
	}

	if user == nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Error during login", utils.FormatValidationErrors(err))
	}

	JWT, err := utils.GenerateToken(int(user.ID), user.Avatar, user.Email, user.Name, user.Token, user.Bio, user.EmailVerified, user.FollowerCount, user.FollowingCount)

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate JWT token", err.Error())
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
			"Bio":        user.Bio,
			"IsVerified": user.EmailVerified,
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
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := ac.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	findUser, err := ac.db.FindUserByEmail(req.Email)

	if err != nil && err != gorm.ErrRecordNotFound {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to check user existence", err.Error())
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

	FindToken, err := ac.db.FindUserByToken(html.EscapeString(Token))

	if err != nil && err != gorm.ErrRecordNotFound {
		// Return an error response if there is an error other than record not found.
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if FindToken == nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	JWT, err := utils.GenerateToken(int(FindToken.ID), FindToken.Avatar, FindToken.Email, FindToken.Name, FindToken.Token, FindToken.Bio, FindToken.EmailVerified, FindToken.FollowerCount, FindToken.FollowingCount)

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate JWT token", err.Error())
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
	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":  "Invalid or missing authentication",
			"status": fiber.StatusUnauthorized,
		})
	}

	ID := c.Params("ID")
	claimsID := strconv.Itoa(claims.UserID)

	if claimsID != ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "You are not authorized to delete this account",
			"status": fiber.StatusBadRequest,
		})
	}

	// Convert string ID to uint
	userID, err := strconv.ParseUint(claimsID, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid user ID format",
			"status": fiber.StatusBadRequest,
			"detail": "User ID must be a positive number",
		})
	}

	deleteUser, err := ac.db.FindUserById(uint(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "User not found",
				"status": fiber.StatusNotFound,
				"detail": "The requested user does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Database error",
			"status": fiber.StatusInternalServerError,
			"detail": err.Error(),
		})
	}

	if deleteUser == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":  "User not found",
			"status": fiber.StatusNotFound,
		})
	}

	// Delete the user from the database first
	Delete, err := ac.db.DeleteUser(claimsID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete user from database",
			"status": fiber.StatusInternalServerError,
			"detail": err.Error(),
			"User":   Delete.Name,
		})
	}

	// Handle avatar deletion in background if it exists
	if deleteUser.Avatar != "" {
		go func() {
			publicID, err := utils.ExtractPublicID(deleteUser.Avatar)
			if err != nil {
				log.Printf("Error extracting public ID: %v", err)
				return
			}

			cld, err := config.InitCloudinary()
			if err != nil {
				log.Printf("Error initializing Cloudinary: %v", err)
				return
			}

			if err := utils.DeleteImageFromCloudinary(cld, publicID); err != nil {
				log.Printf("Error deleting image from Cloudinary: %v", err)
			}
		}()
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User successfully deleted",
		"user": fiber.Map{
			"id":    deleteUser.ID,
			"email": deleteUser.Email,
			"name":  deleteUser.Name,
		},
	})
}

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the End of the DeleteRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the Start of the EditRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

type EditUserRequest struct {
	Name     string `form:"name" validate:"omitempty,max=255"`
	Email    string `form:"email" validate:"omitempty,email,max=255"`
	Password string `form:"password" validate:"omitempty,max=255,min=8"`
	Bio      string `form:"bio" validate:"omitempty,max=255"`
}

func (ac *AuthController) EditUser(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	var req EditUserRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := ac.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	// Get existing user
	existingUser, err := ac.db.FindUserById(uint(claims.UserID))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	// Handle avatar upload if provided
	file, err := c.FormFile("avatar")
	if err == nil {
		uploadChan := make(chan struct {
			url string
			err error
		})

		cld, err := config.InitCloudinary()
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to initialize Cloudinary", err.Error())
		}

		ctx := context.Background()

		go func(file *multipart.FileHeader) {
			var result struct {
				url string
				err error
			}

			if err := utils.ValidateImageFile(file); err != nil {
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

			// Delete old avatar if exists
			if existingUser.Avatar != "" {
				publicID, err := utils.ExtractPublicID(existingUser.Avatar)
				if err == nil {
					utils.DeleteImageFromCloudinary(cld, publicID)
				}
			}

			url, err := utils.UploadToCloudinary(cld, ctx, fileHeader)
			if err != nil {
				result.err = err
				uploadChan <- result
				return
			}

			result.url = url
			uploadChan <- result
		}(file)

		uploadResult := <-uploadChan
		if uploadResult.err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to upload image", uploadResult.err.Error())
		}
		existingUser.Avatar = uploadResult.url
	}

	// Update user fields if provided
	if req.Name != "" {
		existingUser.Name = html.EscapeString(req.Name)
	}
	if req.Email != "" {
		existingUser.Email = html.EscapeString(req.Email)
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
		}
		existingUser.Password = string(hashedPassword)
	}
	if req.Bio != "" {
		existingUser.Bio = html.EscapeString(req.Bio)
	}

	// Update user in database
	updatedUser, err := ac.db.UpdateUser(*existingUser)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User updated successfully",
		"status":  fiber.StatusOK,
		"user": fiber.Map{
			"id":     updatedUser.ID,
			"name":   updatedUser.Name,
			"email":  updatedUser.Email,
			"bio":    updatedUser.Bio,
			"avatar": updatedUser.Avatar,
		},
	})
}

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the End of the EditRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------
