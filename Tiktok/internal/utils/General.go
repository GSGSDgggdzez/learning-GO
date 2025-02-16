package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

func SendErrorResponse(c *fiber.Ctx, status int, message string, details interface{}) error {
	response := fiber.Map{"error": message, "status": status}
	if details != nil {
		response["details"] = details
	}
	return c.Status(status).JSON(response)
}

func GenerateUniqueFilename(original string) string {
	extension := filepath.Ext(original)
	name := strings.TrimSuffix(original, extension)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", name, timestamp, extension)
}

func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		tag := err.Tag()
		errors[field] = formatErrorMessage(field, tag)
	}

	return errors
}

func formatErrorMessage(field, tag string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return "Invalid email format"
	case "min":
		return field + " is too short"
	case "max":
		return field + " is too long"
	case "e164":
		return "Invalid phone number format"
	default:
		return "Invalid " + field
	}
}

func ValidateImageFile(file *multipart.FileHeader) error {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// List of allowed image extensions
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	if !allowedTypes[ext] {
		return fmt.Errorf("invalid file type. Only JPG, JPEG, PNG and GIF are allowed")
	}

	// Open the file for MIME type checking
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Read first 512 bytes to determine MIME type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return err
	}

	// Check MIME type
	contentType := http.DetectContentType(buffer)
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("file is not a valid image")
	}

	return nil
}

func ValidateVideoFile(file *multipart.FileHeader) error {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// List of allowed video extensions
	allowedTypes := map[string]bool{
		".mp4":  true,
		".mov":  true,
		".avi":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
		".mkv":  true,
	}

	if !allowedTypes[ext] {
		return fmt.Errorf("invalid file type. Only MP4, MOV, AVI, WMV, FLV, WEBM, and MKV are allowed")
	}

	// Open the file for MIME type checking
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Read first 512 bytes to determine MIME type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return err
	}

	// Check MIME type
	contentType := http.DetectContentType(buffer)
	validVideoMIME := map[string]bool{
		"video/mp4":        true,
		"video/quicktime":  true,
		"video/x-msvideo":  true,
		"video/x-ms-wmv":   true,
		"video/x-flv":      true,
		"video/webm":       true,
		"video/x-matroska": true,
	}

	if !validVideoMIME[contentType] {
		return fmt.Errorf("file is not a valid video")
	}

	// Check file size (e.g., max 100MB)
	maxSize := int64(100 * 1024 * 1024) // 100MB in bytes
	if file.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum limit of 100MB")
	}

	return nil
}

// 🔄 Helper function to process hashtags (making them Instagram-worthy)
func ProcessHashtags(tags []string) []string {
	// 🧹 Clean up those hashtags like cleaning your room (but actually doing it)
	processedTags := make([]string, 0)
	for _, tag := range tags {
		// Remove any # if present and trim spaces
		cleanTag := strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if cleanTag != "" {
			processedTags = append(processedTags, cleanTag)
		}
	}
	return processedTags
}
