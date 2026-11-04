package youtube

import "time"

type Item struct {
	Snippet struct {
		ChannelId   string `json:"channelId"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Thumbnail   struct {
			MaxResolution struct {
				Url string `json:"url"`
			} `json:"maxres"`
		} `json:"thumbnails"`
	} `json:"snippet"`
	ContentDetails struct {
		VideoId          string    `json:"videoId"`
		VideoPublishedAt time.Time `json:"videoPublishedAt"`
	} `json:"contentDetails"`
}

type VideoResp struct {
	Items []Item `json:"items"`
}
