package youtube

import (
	"fmt"

	"github.com/bytedance/sonic"
)

const aurl string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Mimi.jpeg"
const yturl string = "https://www.youtube.com/watch?v="

type image struct {
	Url string `json:"url"`
}

type embed struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
	Color       int    `json:"color,omitempty"`
	Image       image  `json:"image,omitempty"`
}

type YTDiscord struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Content   string  `json:"content"`
	Embeds    []embed `json:"embeds"`
}

func YTDiscordConvert(item *Item) ([]byte, error) {
	url := yturl + item.ContentDetails.VideoId
	b := YTDiscord{
		Username:  "Mimi Kagari",
		AvatarURL: aurl,
		Content:   fmt.Sprintf("%s", url),
		Embeds: []embed{
			{
				Title:       item.Snippet.Title,
				Description: item.Snippet.Description,
				Url:         url,
				Color:       16711680,
				Image: image{
					Url: item.Snippet.Thumbnail.MaxResolution.Url,
				},
			},
		},
	}

	bodyBytes, err := sonic.Marshal(b)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}
