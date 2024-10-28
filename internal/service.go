package internal

import (
	"context"
	"mime/multipart"

	vision "cloud.google.com/go/vision/apiv1"
	"google.golang.org/api/option"
)

func ClassifyImage(image multipart.File, apiKey string) (string, error) {
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	img, err := vision.NewImageFromReader(image)
	if err != nil {
		return "", err
	}

	annotations, err := client.DetectLabels(ctx, img, nil, 10)
	if err != nil {
		return "", err
	}

	for _, aannotation := range annotations {
		if aannotation.Description == "ingredient" {
			return "ingrdient", nil
		} else if aannotation.Description == "cooked food" {
			return "cooked food", nil
		}
	}
	return "ingredient", nil
}
