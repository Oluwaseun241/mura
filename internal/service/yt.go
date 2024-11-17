package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

func filterRelevantVideos(videos []YouTubeVideo, keywords []string) []YouTubeVideo {
	var filteredVideos []YouTubeVideo
	for _, video := range videos {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(video.Title), keyword) || strings.Contains(strings.ToLower(video.Description), keyword) {
				filteredVideos = append(filteredVideos, video)
				break
			}
		}
	}
	return filteredVideos
}

func scoreVideo(video YouTubeVideo, query string) int {
	score := 0
	if strings.Contains(strings.ToLower(video.Title), query) {
		score += 5
	}
	if strings.Contains(strings.ToLower(video.Description), query) {
		score += 3
	}
	return score
}

func rankVideos(videos []YouTubeVideo, query string) []YouTubeVideo {
	sort.Slice(videos, func(i, j int) bool {
		return scoreVideo(videos[i], query) > scoreVideo(videos[j], query)
	})
	return videos
}

const youtubeSearchURL = "https://www.googleapis.com/youtube/v3/search"

func YoutubeSearch(query string) ([]YouTubeVideo, error) {
	authKey := os.Getenv("GOOGLE_SERVICE_KEY")
	encodedQuery := url.QueryEscape(query)

	apiUrl := fmt.Sprintf(
		"%s?part=snippet&q=%s&key=%s&type=video&videoCategoryId=26&maxResults=5&order=relevance",
		youtubeSearchURL, encodedQuery, authKey,
	)

	var videos []YouTubeVideo
	var err error
	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		videos, err = YoutubeAPICall(apiUrl)
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
func YoutubeAPICall(apiUrl string) ([]YouTubeVideo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to YouTube API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error: %s", string(bodyBytes))
	}

	var ytResponse YoutubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&ytResponse); err != nil {
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
