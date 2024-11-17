package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadImage(fileByte []byte) error {
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return fmt.Errorf("unable to create cloudinary client: %v", err)
	}

	// Validate the file content
	if len(fileByte) == 0 {
		return fmt.Errorf("image file is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fileReader := bytes.NewReader(fileByte)

	upload, err := cld.Upload.Upload(
		ctx,
		fileReader,
		uploader.UploadParams{
			Folder: "tastreesAI",
		},
	)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}
	log.Printf("Image uploaded successfully: URL: %s, PublicID: %s\n", upload.SecureURL, upload.PublicID)
	return nil
}
