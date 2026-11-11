package discord

import (
	"fmt"
	"github.com/bytedance/sonic"
	"strings"
)

const mimi string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Mimi.jpeg"
const sheena string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Sheena.jpeg"
const ari string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Ari.jpeg"
const seiran string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Seiran.jpeg"
const sheenaRoommate string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Sheena_Roommate.jpeg"

const yturl string = "https://www.youtube.com/watch?v="
const twturl string = "https://fxtwitter.com/%s/status/%s"
const mdurl string = "https://mangadex.org/chapter/%s"

func ConvertToDiscordFormat(source string, identifier string) ([]byte, error) {
	var finalUrl string
	var icon string
	var username string
	switch source {
	case "youtube":
		finalUrl = fmt.Sprintf("%s%s", yturl, identifier)
		icon = mimi
		username = "Mimi Kagari"
	case "twitter":
		parts := strings.SplitN(identifier, "_", 2)
		finalUrl = fmt.Sprintf(twturl, parts[0], parts[1])
		icon = sheena
		username = "Sheena Totsuki"
	case "mangadex":
		finalUrl = fmt.Sprintf(mdurl, identifier)
		icon = ari
		username = "Ari Moudou"
	default:
		return nil, fmt.Errorf("invalid source: %s", source)
	}

	b := DiscordFormat{
		Username:  username,
		AvatarURL: icon,
		Content:   finalUrl,
	}

	bodyBytes, err := sonic.Marshal(b)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}
