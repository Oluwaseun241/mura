package client

import (
	"context"
	"log"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var (
	VisionClient *vision.ImageAnnotatorClient
	GeminiClient *genai.Client
)

func Init() {
	ctx := context.Background()
	var err error

	authKey := os.Getenv("GOOGLE_SERVICE_KEY")
	geminiApiKey := os.Getenv("GEMINI_API_KEY")

	if authKey != "" && geminiApiKey != "" {
		VisionClient, err = vision.NewImageAnnotatorClient(ctx, option.WithAPIKey(authKey))
		if err != nil {
			log.Printf("Failed to create vision client: %v", err)
		}

		GeminiClient, err = genai.NewClient(ctx, option.WithAPIKey(geminiApiKey))
		if err != nil {
			log.Printf("Failed to create gemini client: %v", err)
		}
	}

}
