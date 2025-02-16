package config

import (
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/joho/godotenv"
)

func InitCloudinary() (*cloudinary.Cloudinary, error) {
	// Load .env file
	godotenv.Load()

	// Create cloudinary instance
	cld, err := cloudinary.NewFromURL(fmt.Sprintf(
		"cloudinary://%s:%s@%s",
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
		os.Getenv("CLOUDINARY_NAME"),
	))

	if err != nil {
		return nil, err
	}

	// Enable HTTPS
	cld.Config.URL.Secure = true

	return cld, nil
}
