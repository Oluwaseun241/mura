package api

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/Oluwaseun241/mura/internal/service"
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

	// Classify the image concurrently
	imageType, err := service.ClassifyImage(fileBytes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status": false,
			"error":  err.Error(),
		})
	}

	response := map[string]interface{}{
		"status": true,
		"type":   imageType,
	}

	// Run all processes concurrently to save time
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Detect ingredients
	if imageType == "ingredient" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ingredients, err := detectIngredients(fileBytes)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				response["error"] = err.Error()
				response["status"] = false
				return
			}
			response["data"] = ingredients

		}()
	} else if imageType == "cooked food" {
		// Detect food and get recipe
		wg.Add(1)
		go func() {
			defer wg.Done()
			food, err := detectFood(fileBytes)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				response["error"] = err.Error()
				response["status"] = false
				return
			}
			response["data"] = food

		}()

		// Upload image(data collection)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := service.UploadImage(fileBytes)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				response["status"] = false
				response["error"] = err.Error()
			}
		}()

		// YouTube recommendation
		wg.Add(1)
		go func() {
			defer wg.Done()
			yt, err := ytVideoRecommendation(fileBytes)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				response["status"] = false
				response["yt_error"] = err.Error()
			}
			response["yt"] = yt

		}()
	} else if imageType == "invalid item detected" {
		response["status"] = false
		response["error"] = "Invalid item detected...please upload appropriate image"
	}

	// Wait for all goroutines to finish
	wg.Wait()
	return c.JSON(http.StatusOK, response)
}

func IngredientHandler(c echo.Context) error {
	// Parse multiple files from the request
	form, err := c.MultipartForm()
	if err != nil || len(form.File["images"]) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No images uploaded"})
	}

	// process image concurrently
	var wg sync.WaitGroup
	imageChannel := make(chan map[string]interface{}, len(form.File["images"]))

	for _, file := range form.File["images"] {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()
			src, err := file.Open()
			if err != nil {
				imageChannel <- map[string]interface{}{"error": "Failed to open uploaded image"}
				return
			}
			defer src.Close()

			fileBytes, err := io.ReadAll(src)
			if err != nil {
				imageChannel <- map[string]interface{}{"error": "Failed to read uploaded image"}
				return
			}

			// Detect ingredients from the image
			ingredientsMap, err := detectIngredients(fileBytes)
			if err != nil {
				imageChannel <- map[string]interface{}{"status": false, "error": err.Error()}
				return
			}

			if foods, ok := ingredientsMap["foods"].([]interface{}); ok {
				imageChannel <- map[string]interface{}{"status": true, "data": foods}
			} else {
				imageChannel <- map[string]interface{}{"status": false, "error": "No ingredients detected"}
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(imageChannel)
	}()

	var allIngredients []interface{}
	for res := range imageChannel {
		if res["status"].(bool) {
			if data, ok := res["data"].([]interface{}); ok {
				allIngredients = append(allIngredients, data...)
			}
		}
	}

	// Remove duplicate
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
	query := fmt.Sprintf("How to make %s", data.Dish)
	yt, err := service.YoutubeSearch(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   recipe,
		"yt":     yt,
	})
}
