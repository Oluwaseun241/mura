package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Oluwaseun241/mura/cmd/client"
	"github.com/google/generative-ai-go/genai"
)

func getFoodRecipes(ingredients []string, dish string) (string, error) {
	ctx := context.Background()

	prompt1 := fmt.Sprintf("You are a helpful, AI assistant devoted to providing accurate and delightful recipes.These are the guildlines for you to follow when delivering a recipe response to a request 1. List out the ingredients first,including quantities.Provide detailed cooking times, temperatures and any special kitchen equipment needed 2.Provide step-by-step instructions for prepping, mixing, cooking, plating and any other necessary steps, detailed enough to follow. Include Pro chef tips and special techniques as applicable Here are the ingredients I have: %s. Can you give me a specific recipe that includes only these ingredients, and detailed preparation steps? Lastly Nutritional information like Calories, Protein and Carbs", strings.Join(ingredients, ", "))
	prompt2 := fmt.Sprintf("You are a helpful, AI assistant devoted to providing accurate and delightful recipes.These are the guildlines for you to follow when delivering a recipe response to a request 1. List out the ingredients first,including quantities.Provide detailed cooking times, temperatures and any special kitchen equipment needed 2.Provide step-by-step instructions for prepping, mixing, cooking, plating and any other necessary steps, detailed enough to follow. Include Pro chef tips and special techniques as applicable Here are the ingredients I have: %s. Can you give me a specific recipe that includes only these ingredients, Nutritional information like Calories, Protein and Carbs, and detailed preparation steps for %s", strings.Join(ingredients, ", "), dish)

	// S.elect the appropriate prompt
	var prompt string
	if dish != "" {
		prompt = prompt2
	} else {
		prompt = prompt1
	}

	model := client.GeminiClient.GenerativeModel("gemini-1.5-pro")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("Error generating content: %v", err)
	}
	if resp == nil {
		return "", fmt.Errorf("no response received from Gemini API")
	}
	return printResponse(resp), nil
}

func detectFood(fileBytes []byte) (string, error) {
	ctx := context.Background()

	prompt := []genai.Part{
		genai.ImageData("jpeg", fileBytes),
		genai.Text("Accurately identify the food in the image and provide an appropriate recipe consistent with your analysis.These are the guildlines for you to follow when delivering a recipe response to a request 1. List out the ingredients first, including quantities. Provide detailed cooking times, temperatures and any special kitchen equipment needed 2.Provide step-by-step instructions for prepping, mixing, cooking, plating and any other necessary steps, detailed enough to follow. Include pro chef tips and special techniques as applicable. Lastly Nutritional information like Calories, Protein and Carbs"),
	}

	model := client.GeminiClient.GenerativeModel("gemini-1.5-pro")
	//model.ResponseMIMEType = "application/json"

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", fmt.Errorf("Error generating content")
	}
	return printResponse(resp), nil
}

func detectIngredients(file []byte) (map[string]interface{}, error) {
	model := client.GeminiClient.GenerativeModel("gemini-1.5-pro")
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

	if foods, ok := parsedResponse["foods"].([]interface{}); ok {
		uniqueFoods := removeDuplicates(foods)
		parsedResponse["foods"] = uniqueFoods
	}

	return parsedResponse, nil
}

func getVideoPrompt(file []byte) (string, error) {
	ctx := context.Background()

	prompt := []genai.Part{
		genai.ImageData("jpeg", file),
		genai.Text("Accurately identify the food in the image and provide an appropriate prompt to search for tutorial video on youtube"),
	}

	model := client.GeminiClient.GenerativeModel("gemini-1.5-pro")
	model.ResponseMIMEType = "application/json"

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", fmt.Errorf("Error generating content")
	}
	return printResponse(resp), nil

}
