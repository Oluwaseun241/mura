package api

import (
	"context"
	"fmt"
	"os"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func printResponse(resp *genai.GenerateContentResponse) string {
	var result strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				result.WriteString(fmt.Sprintf("%v\n", part))
			}
		}
	}
	return result.String()
}

func getFoodRecipes(ingredients []string, geminiApiKey string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiApiKey))
	if err != nil {
		return "", fmt.Errorf("Error initializing Gemini API: %v", err)
	}

	prompt := fmt.Sprintf("Here are the ingredients I have: %s. Can you give me a specific recipe that includes only these ingredients, and detailed preparation steps?", strings.Join(ingredients, ", "))
	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("Error generating content: %v", err)
	}

	if resp == nil {
		return "", fmt.Errorf("no response received from Gemini API")
	}
	return printResponse(resp), nil
}

func detectIngredients(imagePath string, visionClient *vision.ImageAnnotatorClient) ([]string, error) {
	// Read the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read image file: %v", err)
	}
	defer file.Close()

	image, err := vision.NewImageFromReader(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert image to vision format: %v", err)
	}

	// Perform label detection
	labels, err := visionClient.LocalizeObjects(context.Background(), image, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to detect labels: %v", err)
	}

	// Collect ingredients, ensuring no duplicates
	ingredientMap := make(map[string]bool)
	var ingredients []string
	for _, label := range labels {
		if label.Score > 0.55 {
			ingredient := label.Name
			if !ingredientMap[ingredient] {
				ingredientMap[ingredient] = true
				ingredients = append(ingredients, ingredient)
			}
		}
	}
	return ingredients, nil
}
