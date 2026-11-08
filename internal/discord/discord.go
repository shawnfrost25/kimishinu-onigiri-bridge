package discord

import (
	"net/http"
)

type Discord struct {
	HttpClient http.Client
	WebhookURL string
}

func New(httpClient http.Client, webhookURL string) *Discord {
	return &Discord{
		HttpClient: httpClient,
		WebhookURL: webhookURL,
	}
}
