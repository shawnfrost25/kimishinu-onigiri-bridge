package discord

import (
	"log/slog"
	"net/http"
)

type Discord struct {
	HttpClient http.Client
	WebhookURL string
	Logger     slog.Logger
}

func New(httpClient http.Client, webhookURL string, logger slog.Logger) *Discord {
	return &Discord{
		HttpClient: httpClient,
		WebhookURL: webhookURL,
		Logger:     logger,
	}
}
