package internal

import (
	"bytes"
	"context"

	vision "cloud.google.com/go/vision/apiv1"
	"google.golang.org/api/option"
)

func ClassifyImage(imageBytes []byte, cred string) (string, error) {
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithAPIKey(cred))
	if err != nil {
		return "", err
	}
	defer client.Close()

	imageReader := bytes.NewReader(imageBytes)

	img, err := vision.NewImageFromReader(imageReader)
	if err != nil {
		return "", err
	}

	annotations, err := client.LocalizeObjects(ctx, img, nil)
	if err != nil {
		return "", err
	}

	cookedFoodTerms := []string{"Food", "Recipe", "Cuisine", "Dish", "Jollof rice", "Fried rice", "Rice"}
	ingredientTerms := []string{"Ingredient", "Vegetable", "Spice"}

	for _, annotation := range annotations {
		for _, term := range cookedFoodTerms {
			if annotation.Name == term && annotation.Score >= 0.50 {
				return "cooked food", nil
			}
		}
		for _, term := range ingredientTerms {
			if annotation.Name == term && annotation.Score >= 0.50 {
				return "ingredient", nil
			}
		}
	}
	return "invalid item detected", nil
}

// Returns youtube video relating to the recipe
func YoutubeSearch() {}
