package service

type VideoPromptResponse struct {
	FoodName            string `json:"food_name"`
	YouTubeSearchPrompt string `json:"youtube_search_prompt"`
}

type YoutubeResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title        string `json:"title"`
			Description  string `json:"description"`
			ChannelTitle string `json:"channelTitle"`
			Thumbnails   struct {
				High struct {
					URL string `json:"url"`
				} `json:"high"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}

type YouTubeVideo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	VideoURL    string `json:"videoUrl"`
}
