package api

import (
	"io"
	"net/http"

	"github.com/Oluwaseun241/mura/internal"
	"github.com/labstack/echo/v4"
)

func FoodHandler(c echo.Context) error {
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
	imageType, err := internal.ClassifyImage(fileBytes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status": false,
			"error":  err.Error(),
		})
	}

	if imageType == "ingredient" {
		ingredients, err := detectIngredients(fileBytes)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": true,
			"type":   "ingredient",
			"data":   ingredients,
		})
	} else if imageType == "cooked food" {
		food, err := detectFood(fileBytes)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status": false,
				"error":  err.Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": true,
			"type":   "food",
			"data":   food,
		})
	} else if imageType == "invalid item detected" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": false,
			"error":  "invalid item detected...please upload appropriate image",
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": false,
		"error":  "Unknown image type",
	})
}

func IngredientHandler(c echo.Context) error {
	// Parse multiple files from the request
	form, err := c.MultipartForm()
	if err != nil || len(form.File["images"]) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No images uploaded"})
	}

	allIngredients := []interface{}{}

	for _, file := range form.File["images"] {
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
		ingredientsMap, err := detectIngredients(fileBytes)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status": false,
				"error":  err.Error(),
			})
		}

		if foods, ok := ingredientsMap["foods"].([]interface{}); ok {
			allIngredients = append(allIngredients, foods...)
		}
	}

	uniqueIngredients := removeDuplicates(allIngredients)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   uniqueIngredients,
	})
}

func RecipeHandler(c echo.Context) error {
	var data struct {
		Ingredients []string `json:"ingredients"`
		Dish        string   `json:"dish"`
	}

	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No ingredients provided"})
	}

	//Get food recipes using detected ingredients from Gemini API
	recipe, err := getFoodRecipes(data.Ingredients, data.Dish)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   recipe,
	})
}

func YtHandler(c echo.Context) error {
	query := "HowtocookPasta"
	video, err := internal.YoutubeSearch(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   video,
	})
}
