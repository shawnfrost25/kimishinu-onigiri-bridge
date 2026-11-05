package youtube

import "time"

type Item struct {
	Snippet struct {
		ChannelId   string `json:"channelId"`
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"snippet"`
	ContentDetails struct {
		VideoId          string    `json:"videoId"`
		VideoPublishedAt time.Time `json:"videoPublishedAt"`
	} `json:"contentDetails"`
}

type VideoResp struct {
	Items []Item `json:"items"`
}
