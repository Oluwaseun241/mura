package client

import (
	"context"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var (
	GeminiClient *genai.Client
)

func Init() {
	ctx := context.Background()
	var err error

	geminiApiKey := os.Getenv("GEMINI_API_KEY")

	if geminiApiKey != "" {
		GeminiClient, err = genai.NewClient(ctx, option.WithAPIKey(geminiApiKey))
		if err != nil {
			log.Printf("Failed to create gemini client: %v", err)
		}
	}
}
