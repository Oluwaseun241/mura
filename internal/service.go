package internal

import (
	"bytes"
	"context"

	vision "cloud.google.com/go/vision/apiv1"
	"google.golang.org/api/option"
)

func ClassifyImage(imageBytes []byte, cred string) (string, error) {
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentialsFile(cred))
	if err != nil {
		return "", err
	}
	defer client.Close()

	imageReader := bytes.NewReader(imageBytes)

	img, err := vision.NewImageFromReader(imageReader)
	if err != nil {
		return "", err
	}

	annotations, err := client.DetectLabels(ctx, img, nil, 10)
	if err != nil {
		return "", err
	}

	cookedFoodTerms := []string{"Food", "Recipe", "Cuisine", "Dish", "Jollof rice", "Fried rice", "Rice"}
	ingredientTerms := []string{"Ingredient", "Vegetable", "Spice"}

	for _, annotation := range annotations {
		for _, term := range cookedFoodTerms {
			if annotation.Description == term && annotation.Score >= 0.75 {
				return "cooked food", nil
			}
		}
		for _, term := range ingredientTerms {
			if annotation.Description == term && annotation.Score >= 0.75 {
				return "ingredient", nil
			}
		}
	}
	return "ingredient", nil
}
