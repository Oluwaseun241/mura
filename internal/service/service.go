package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Oluwaseun241/mura/cmd/client"
	"github.com/google/generative-ai-go/genai"
)

func ClassifyImage(imageBytes []byte) (string, error) {
	ctx := context.Background()

	prompt := []genai.Part{
		genai.ImageData("jpeg", imageBytes),
		genai.Text("Analyze this image and classify it as either 'cooked food' or 'ingredient'. Return the result in JSON format as {\"type\": \"cooked food\"} or {\"type\": \"ingredient\"}. If the image doesn't contain food or ingredients, return {\"type\": \"invalid\"}."),
	}

	model := client.GeminiClient.GenerativeModel("gemini-2.0-flash")
	model.ResponseMIMEType = "application/json"

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}

	// Extract the content from the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	var combinedContent string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			combinedContent += string(textPart)
		} else {
			return "", fmt.Errorf("unexpected part type: %T", part)
		}
	}

	var result struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(combinedContent), &result); err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	return result.Type, nil
}
