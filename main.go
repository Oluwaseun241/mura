package main

import (
	"context"
	"fmt"
	"log"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// Load Env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS environment variable is not set.")
	}

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
	labels, err := client.LocalizeObjects(ctx, image, nil)
	if err != nil {
		log.Fatalf("Failed to detect labels: %v", err)
	}

	// Print the detected labels
	fmt.Println("Detected labels:")
	for _, label := range labels {
		if label.Score <= 0.59 {
			fmt.Printf("Not sure of %s", label.Name)
		}
		fmt.Printf("%s (Confidence: %f)\n", label.Name, label.Score)
	}
}
