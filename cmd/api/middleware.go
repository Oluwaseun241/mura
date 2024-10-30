package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

func parseResponse(resp string) (string, error) {
	var ingredients struct {
		Ingredients []string `json:"ingredients"`
	}

	// Use json.Unmarshal to parse the structured text response into JSON
	err := json.Unmarshal([]byte(resp), &ingredients)
	if err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %v", err)
	}

	// Format it back to a string to return as JSON
	formattedResponse, err := json.MarshalIndent(ingredients, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response JSON: %v", err)
	}

	return string(formattedResponse), nil
}

func getFoodRecipes(ingredients []string, dish string, geminiApiKey string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiApiKey))
	if err != nil {
		return "", fmt.Errorf("Error initializing Gemini API: %v", err)
	}

	prompt1 := fmt.Sprintf("You are a helpful, AI assistant devoted to providing accurate and delightful recipes.This are the guildlines for you to follow when delivering a recipe response to a request 1. List out the ingredients first,including quantities.Provide detailed cooking times, temperatures and any special kitchen equipment needed 2.Provide step-by-step instructions for prepping, mixing, cooking, plating and any other necessary steps, detailed enough to follow. Include safety tips and special techniques as applicable Here are the ingredients I have: %s. Can you give me a specific recipe that includes only these ingredients, and detailed preparation steps?", strings.Join(ingredients, ", "))
	prompt2 := fmt.Sprintf("You are a helpful, AI assistant devoted to providing accurate and delightful recipes.This are the guildlines for you to follow when delivering a recipe response to a request 1. List out the ingredients first,including quantities.Provide detailed cooking times, temperatures and any special kitchen equipment needed 2.Provide step-by-step instructions for prepping, mixing, cooking, plating and any other necessary steps, detailed enough to follow. Include safety tips and special techniques as applicable Here are the ingredients I have: %s. Can you give me a specific recipe that includes only these ingredients, and detailed preparation steps for %s", strings.Join(ingredients, ", "), dish)

	// Select the appropriate prompt
	var prompt string
	if dish != "" {
		prompt = prompt2
	} else {
		prompt = prompt1
	}

	model := client.GenerativeModel("gemini-1.5-pro")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("Error generating content: %v", err)
	}
	if resp == nil {
		return "", fmt.Errorf("no response received from Gemini API")
	}
	return printResponse(resp), nil
}

func detectFood(fileBytes []byte, geminiApiKey string) (string, error) {
	ctx := context.Background()
	prompt := []genai.Part{
		genai.ImageData("jpeg", fileBytes),
		genai.Text("Accurately identify the food in the image and provide an appropriate recipe consistent with your analysis."),
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiApiKey))
	if err != nil {
		return "", fmt.Errorf("Error initializing Gemini API")
	}

	model := client.GenerativeModel("gemini-1.5-pro")
	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", fmt.Errorf("Error generating content")
	}
	return printResponse(resp), nil
}

func detectIngredients(file []byte, apiKey string) (map[string]interface{}, error) {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("Error initializing Gemini API")
	}
	model := client.GenerativeModel("gemini-1.5-pro")
	model.ResponseMIMEType = "application/json"
	prompt := []genai.Part{
		genai.ImageData("jpeg", file),
		genai.Text("Identify and list all food items in this image with accurate labels in JSON format. Please return the result as a valid JSON object formatted as {'foods': ['item1', 'item2', ...]} without any additional text, comments, or formatting issues."),
	}
	resp, err := model.GenerateContent(context.Background(), prompt...)
	if err != nil {
		return nil, fmt.Errorf("error generating content: %v", err)
	}

	// Extract the content from the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	var combinedContent string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			combinedContent += string(textPart)
		} else {
			return nil, fmt.Errorf("unexpected part type: %T", part)
		}
	}

	var parsedResponse map[string]interface{}
	err = json.Unmarshal([]byte(combinedContent), &parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return parsedResponse, nil
}

// Check if they are food item
func validateIngredient(ingredients []string) ([]string, []string, error) {
	validIngredients := []string{}
	invalidIngredients := []string{}

	// open and read txt file content
	input, err := os.Open("./internal/data.txt")
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
