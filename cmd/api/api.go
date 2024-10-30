package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/Oluwaseun241/mura/internal"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
)

func FoodHandler(c echo.Context) error {
	geminiApiKey := os.Getenv("GEMINI_API_KEY")

	authKey := os.Getenv("GOOGLE_SERVICE_KEY")

	visionClient, err := vision.NewImageAnnotatorClient(context.Background(), option.WithAPIKey(authKey))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create Vision client"})
	}
	defer visionClient.Close()

	// Parse multipart form data
	form, err := c.MultipartForm()
	if err != nil || len(form.File["image"]) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No images uploaded"})
	}

	file := form.File["image"][0]

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open uploaded image"})
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read uploaded image"})
	}

	// Classify the image
	imageType, err := internal.ClassifyImage(fileBytes, authKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	if imageType == "ingredient" {
		ingredients, err := detectIngredients(fileBytes, geminiApiKey)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"status": true, "data": ingredients})
	} else if imageType == "cooked food" {
		food, err := detectFood(fileBytes, geminiApiKey)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": true,
			"data":   food,
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": false,
		"error":  "Unknown image type",
	})
}

func IngredientHandler(c echo.Context) error {
	geminiApiKey := os.Getenv("GEMINI_API_KEY")

	authKey := os.Getenv("GOOGLE_SERVICE_KEY")

	// if credsPath == "" {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Credentials missing"})
	// }

	visionClient, err := vision.NewImageAnnotatorClient(context.Background(), option.WithAPIKey(authKey))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create Vision client"})
	}
	defer visionClient.Close()

	// Parse multiple files from the request
	form, err := c.MultipartForm()
	if err != nil || len(form.File["images"]) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No images uploaded"})
	}

	//allIngredients := []string{}

	for _, file := range form.File["image"] {
		// Save the uploaded file temporarily
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open uploaded image"})
		}
		defer src.Close()

		fileBytes, err := io.ReadAll(src)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read uploaded image"})
		}
		// Detect ingredients from the image
		ingredients, err := detectIngredients(fileBytes, geminiApiKey)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		}

		fmt.Println(ingredients)
		//allIngredients = append(allIngredients, ingredients...)
	}

	//Remove duplicates from the ingredient list
	//response := removeDuplicates(allIngredients)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   "yoo",
	})
}

func RecipeHandler(c echo.Context) error {
	geminiApiKey := os.Getenv("GEMINI_API_KEY")

	if geminiApiKey == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Credentials missing"})
	}

	var data struct {
		Ingredients []string `json:"ingredients"`
		Dish        string   `json:"dish"`
	}

	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No ingredients provided"})
	}

	//Get food recipes using detected ingredients from Gemini API
	recipe, err := getFoodRecipes(data.Ingredients, data.Dish, geminiApiKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   recipe,
	})
}
