package config

import (
	"os"

	"github.com/GetStream/getstream-go"
)

func InitStreamVideo() (*getstream.Stream, error) {
	apiKey := os.Getenv("STREAM_API_KEY")
	apiSecret := os.Getenv("STREAM_API_SECRET")

	client, err := getstream.NewClient(apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	return client, nil
}
