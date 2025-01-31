package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-playground/validator"
)

func GenerateUniqueFilename(original string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d_%s", timestamp, original)
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
