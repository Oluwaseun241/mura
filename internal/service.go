package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	vision "cloud.google.com/go/vision/apiv1"
	"google.golang.org/api/option"
)

func ClassifyImage(imageBytes []byte) (string, error) {
	authKey := os.Getenv("GOOGLE_SERVICE_KEY")

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithAPIKey(authKey))
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

type YoutubeResoonse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Thumbnails  struct {
				High struct {
					URL string `json:"url"`
				} `json:"default"`
			} `json:"thumbnails"`
			ChannelTitle string `json:"channelTitle"`
		} `json:"snippet"`
	} `json:"items"`
}

type YouTubeVideo struct {
	VideoID     string `json:"videoId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	VideoURL    string `json:"videoUrl"`
}

const youtubeSearchURL = "https://www.googleapis.com/youtube/v3/search"

// Returns youtube video relating to the recipe
func YoutubeSearch(query string) ([]YouTubeVideo, error) {
	authKey := os.Getenv("GOOGLE_SERVICE_KEY")
	client := &http.Client{Timeout: 10 * time.Second}
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

	var ytResponse YoutubeResoonse
	if err := json.Unmarshal(body, &ytResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling YouTube API response: %v", err)
	}

	videos := []YouTubeVideo{}
	for _, item := range ytResponse.Items {
		videos = append(videos, YouTubeVideo{
			VideoID:     item.ID.VideoID,
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			Thumbnail:   item.Snippet.Thumbnails.High.URL,
			VideoURL:    "https://www.youtube.com/watch?v=" + item.ID.VideoID,
		})
	}
	return videos, nil
}
