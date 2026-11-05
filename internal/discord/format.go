package discord

import (
	"fmt"

	"github.com/bytedance/sonic"
)

const mimi string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Mimi.jpeg"
const sheena string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Sheena.jpeg"
const ari string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Ari.jpeg"
const seiran string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Seiran.jpeg"
const sheenaRoommate string = "https://cdn.jsdelivr.net/gh/shawnfrost25/kimishinu-onigiri-bridge@main/assets/Sheena_Roommate.jpeg"

const yturl string = "https://www.youtube.com/watch?v="

func ConvertToDiscordFormat(source string, identifier string) ([]byte, error) {
	var finalUrl string
	var icon string
	var username string
	switch source {
	case "youtube":
		finalUrl = fmt.Sprintf("Mimi found a new **KimiShinu** drop. Don't pretend you weren't waiting for this. Just copy and share with others!!\n\n%s%s", yturl, identifier)
		icon = mimi
		username = "Mimi Kagari"
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
