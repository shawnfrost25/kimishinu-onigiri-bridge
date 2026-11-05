package discord

type Image struct {
	Url string `json:"url"`
}

type Embed struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
	Color       int    `json:"color,omitempty"`
	Image       Image  `json:"image,omitempty"`
}

type DiscordFormat struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Content   string  `json:"content"`
	Embeds    []Embed `json:"embeds"`
}
