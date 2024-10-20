package api

import (
	"bufio"
	"context"
	"fmt"
	"mime/multipart"
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

func getFoodRecipes(ingredients []string, dish string, geminiApiKey string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiApiKey))
	if err != nil {
		return "", fmt.Errorf("Error initializing Gemini API: %v", err)
	}

	prompt1 := fmt.Sprintf("Here are the ingredients I have: %s. Can you give me a specific recipe that includes only these ingredients, and detailed preparation steps?", strings.Join(ingredients, ", "))
	prompt2 := fmt.Sprintf("Here are the ingredients I have: %s. Can you give me a specific recipe that includes only these ingredients, and detailed preparation steps for %s", strings.Join(ingredients, ", "), dish)

	// Select the appropriate prompt
	var prompt string
	if dish != "" {
		prompt = prompt2
	} else {
		prompt = prompt1
	}

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

func detectIngredients(file multipart.File, visionClient *vision.ImageAnnotatorClient) ([]string, error) {
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

	// validate detected ingredients
	valid, invalid, err := validateIngredient(ingredients)
	if err != nil {
		return nil, fmt.Errorf("Error: %s", err)
	}
	if len(valid) == 0 && len(invalid) > 0 {
		return nil, fmt.Errorf("Invalid food item: found %v", invalid)
	}
	return valid, nil
}

// Check if they are food item
func validateIngredient(ingredients []string) ([]string, []string, error) {
	validIngredients := []string{}
	invalidIngredients := []string{}

	// open and read txt file content
	input, err := os.Open("./tmp/data.txt")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer input.Close()

	// check through a text file
	validSet := make(map[string]bool)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		validSet[strings.TrimSpace(scanner.Text())] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading file: %v", err)
	}
	for _, ingredient := range ingredients {
		if validSet[strings.ToLower(strings.TrimSpace(ingredient))] {
			validIngredients = append(validIngredients, ingredient)
		} else {
			invalidIngredients = append(invalidIngredients, ingredient)
		}
	}
	return validIngredients, invalidIngredients, nil
}

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for _, v := range elements {
		if !encountered[v] {
			encountered[v] = true
			result = append(result, v)
		}
	}

	return result
}
