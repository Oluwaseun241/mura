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

	for _, aannotation := range annotations {
		if aannotation.Description == "ingredient" {
			return "ingrdient", nil
		} else if aannotation.Description == "cooked food" {
			return "cooked food", nil
		}
	}
	return "ingredient", nil
}
