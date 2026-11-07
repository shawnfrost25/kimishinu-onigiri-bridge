package lib

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

type Author struct {
	UserName string `json:"userName"`
}

type Tweet struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Text   string `json:"text"`
	Author Author `json:"author"`
}

type TWTData struct {
	Tweets []Tweet `json:"tweets"`
}

type TwitterResp struct {
	Data TWTData `json:"data"`
}

type MDData struct {
	ID string `json:"id"`
}

type MangaDexResp struct {
	Result string   `json:"result"`
	Data   []MDData `json:"data"`
}
