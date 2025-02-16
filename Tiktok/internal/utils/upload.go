package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(cld *cloudinary.Cloudinary, ctx context.Context, file interface{}) (string, error) {
	resp, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		UniqueFilename: api.Bool(true),
		Folder:         "tiktok-clone",
		ResourceType:   "video",
		Transformation: "w_1280,q_auto:good,vc_auto,f_auto",
	})

	if err != nil {
		return "", err
	}

	return resp.SecureURL, nil
}
func DeleteImageFromCloudinary(cld *cloudinary.Cloudinary, publicID string) error {
	// Attempt to delete the image from Cloudinary
	deleteResp, err := cld.Upload.Destroy(context.Background(), uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete image from Cloudinary: %v", err)
	}

	// Check the Cloudinary response
	if deleteResp.Result != "ok" {
		return fmt.Errorf("failed to delete image from Cloudinary, response: %v", deleteResp)
	}

	return nil
}

func ExtractPublicID(url string) (string, error) {
	// Define the base URL part that we need to remove
	baseURL := "https://res.cloudinary.com/"

	// Check if the URL starts with the Cloudinary base URL
	if !strings.HasPrefix(url, baseURL) {
		return "", fmt.Errorf("invalid Cloudinary URL")
	}

	// Remove the base URL part
	urlWithoutBase := strings.TrimPrefix(url, baseURL)

	// The next segment should be the image upload part, remove this
	uploadSegment := "/image/upload/"
	urlWithoutUploadSegment := strings.Split(urlWithoutBase, uploadSegment)[1]

	// Remove the file extension (e.g., ".jpg", ".png", etc.)
	// We can assume the extension will be the part after the last period (.)
	parts := strings.Split(urlWithoutUploadSegment, ".")
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid Cloudinary URL format")
	}

	// The first part of the split is the public ID
	publicID := parts[0]
	return publicID, nil
}
