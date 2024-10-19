package api

import (
	"context"
	"net/http"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
)

// TODO: Support multiple image upload
//
//	Ability to specify food you have in mind(later)
func IngredientHandler(c echo.Context) error {
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if credsPath == "" {
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

	return c.JSON(http.StatusOK, map[string]interface{}{
		"ingredients": ingredients,
	})
}

func RecipeHandler(c echo.Context) error {
	geminiApiKey := os.Getenv("GEMINI_API_KEY")

	if geminiApiKey == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Credentials missing"})
	}

	var data struct {
		Ingredients []string `json:"ingredients"`
	}

	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No ingredients provided"})
	}

	//Get food recipes using detected ingredients from Gemini API
	recipe, err := getFoodRecipes(data.Ingredients, geminiApiKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"recipe": recipe,
	})
}
