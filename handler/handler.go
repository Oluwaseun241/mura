package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/google/generative-ai-go/genai"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
)

// TODO: return recipe steps
func printResponse(resp *genai.GenerateContentResponse) string {
	var result strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			recipeFound := false
			for _, part := range cand.Content.Parts {
				if recipeFound {
					result.WriteString(fmt.Sprintf("%v\n", part))
				}
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

func RecipeHandler(c echo.Context) error {

	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	geminiApiKey := os.Getenv("GEMINI_API_KEY")

	if credsPath == "" || geminiApiKey == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Credentials missing"})
	}

	visionClient, err := vision.NewImageAnnotatorClient(context.Background(), option.WithCredentialsFile(credsPath))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create Vision client"})
	}
	defer visionClient.Close()

	// Get image file from the request
	file, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No image uploaded"})
	}

	// Save the uploaded file temporarily
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open uploaded image"})
	}
	defer src.Close()

	// Save the file to disk
	imagePath := "./tmp/uploaded_image.jpeg"
	dst, err := os.Create(imagePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save image"})
	}
	defer dst.Close()

	// Copy the uploaded image to the destination file
	if _, err = dst.ReadFrom(src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save image"})
	}

	// Detect ingredients from the image
	ingredients, err := detectIngredients(imagePath, visionClient)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	fmt.Println(ingredients)
	//Get food recipes from Gemini API
	recipe, err := getFoodRecipes(ingredients, geminiApiKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	fmt.Println(recipe)
	return c.JSON(http.StatusOK, map[string]string{
		"recipe": "yoo",
	})
}
