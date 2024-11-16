package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/Oluwaseun241/mura/cmd/client"
)

func ClassifyImage(imageBytes []byte) (string, error) {
	ctx := context.Background()
	imageReader := bytes.NewReader(imageBytes)

	img, err := vision.NewImageFromReader(imageReader)
	if err != nil {
		return "", err
	}

	annotations, err := client.VisionClient.LocalizeObjects(ctx, img, nil)
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

const youtubeSearchURL = "https://www.googleapis.com/youtube/v3/search"

func YoutubeSearch(query string) ([]YouTubeVideo, error) {
	authKey := os.Getenv("GOOGLE_SERVICE_KEY")
	encodedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("%s?part=snippet&q=%s&key=%s&type=video&maxResults=5&order=relevance", youtubeSearchURL, encodedQuery, authKey)

	var videos []YouTubeVideo
	var err error
	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		videos, err = YoutubeAPICall(url)
		if err == nil {
			return videos, nil
		}
		// Retry only if the error is related to timeout
		if !errors.Is(err, context.DeadlineExceeded) {
			break
		}
		time.Sleep(2 * time.Second) // wait before retrying
	}

	return nil, fmt.Errorf("YouTube API request failed after %d attempts: %v", maxAttempts, err)
}

// Returns youtube video relating to the recipe
func YoutubeAPICall(query string) ([]YouTubeVideo, error) {
	authKey := os.Getenv("GOOGLE_SERVICE_KEY")
	client := &http.Client{Timeout: 3 * time.Second}
	url := fmt.Sprintf("%s?part=snippet&q=%s&key=%s&type=video&maxResults=5&order=viewCount", youtubeSearchURL, query, authKey)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error sending request to YouTube API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error: %s", string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ytResponse YoutubeResponse
	if err := json.Unmarshal(body, &ytResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling YouTube API response: %v", err)
	}

	videos := []YouTubeVideo{}
	for _, item := range ytResponse.Items {
		videos = append(videos, YouTubeVideo{
			//VideoID:   item.ID.VideoID,
			Title:     item.Snippet.Title,
			Thumbnail: item.Snippet.Thumbnails.High.URL,
			VideoURL:  "https://www.youtube.com/watch?v=" + item.ID.VideoID,
		})
	}
	return videos, nil
}
