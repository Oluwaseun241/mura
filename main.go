package main

import (
	"context"
	"fmt"
	"log"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// Initialize a Vision client with your credentials
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		log.Fatalf("Failed to create Vision client: %v", err)
	}
	defer client.Close()

	// Read the image file
	imagePath, err := os.Open("./raw.jpeg")
	if err != nil {
		log.Fatalf("Failed to read image path: %v", err)
	}
	defer imagePath.Close()

	// Convert the image file to Google Vision's Image type
	image, _ := vision.NewImageFromReader(imagePath)

	// Perform label detection on the image
	labels, err := client.DetectLabels(ctx, image, nil, 10)
	if err != nil {
		log.Fatalf("Failed to detect labels: %v", err)
	}

	// Print the detected labels
	fmt.Println("Detected labels:")
	for _, label := range labels {
		fmt.Printf("%s (Confidence: %f)\n", label.Description, label.Score)
	}
}
