package youtube

import (
	db "kimishinu/internal/db/sqlc"
	"net/http"
)

type Youtube struct {
	ApiKey     string
	WebhookURL string
	Queries    *db.Queries
	HttpClient http.Client
}

func New(apiKey string, webhookURL string, queries *db.Queries, httpClient http.Client) *Youtube {
	return &Youtube{
		ApiKey:     apiKey,
		WebhookURL: webhookURL,
		Queries:    queries,
		HttpClient: httpClient,
	}
}
